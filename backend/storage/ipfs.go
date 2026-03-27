// Package storage provides content-addressed storage for evaluation reports.
//
// Backends supported (in priority order):
//   - LighthouseClient: Lighthouse.storage — stable Filecoin/IPFS pinning service (primary)
//   - FilecoinClient: go-synapse SDK — native Filecoin PDP storage (fallback, when providers register)
//   - BeryxClient: Beryx (Zondax) — authenticated Filecoin RPC for chain queries
//   - IPFSClient: Local IPFS node via HTTP API (development)
//
// All storage backends implement the Store interface for swappable use.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/data-preservation-programs/go-synapse"
	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/spregistry"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	gocid "github.com/ipfs/go-cid"
)

// Store defines the interface for content-addressed storage backends.
// Upload stores a JSON report and returns its CID.
// Retrieve fetches a report by its CID.
type Store interface {
	Upload(report []byte) (string, error)
	Retrieve(cid string) ([]byte, error)
}

// --- Beryx Client (Zondax authenticated Filecoin RPC) ---

// BeryxClient provides authenticated access to the Filecoin network
// via the Beryx (Zondax) RPC proxy. It wraps go-ethereum's ethclient
// with Bearer token auth for Beryx's API.
type BeryxClient struct {
	ethClient *ethclient.Client
	rpcClient *rpc.Client
	chainID   *big.Int
	ctx       context.Context
}

// BeryxConfig holds the configuration for connecting to Beryx.
type BeryxConfig struct {
	// RPCURL is the Beryx RPC proxy endpoint.
	// Mainnet: https://api.zondax.ch/fil/node/mainnet/rpc/v1
	// Calibration: https://api.zondax.ch/fil/node/calibration/rpc/v1
	RPCURL string
	// APIToken is the Beryx JWT token for authentication.
	APIToken string
}

// NewBeryxClient creates a Filecoin client connected via Beryx RPC.
// Verifies the connection by fetching the chain ID.
func NewBeryxClient(cfg BeryxConfig) (*BeryxClient, error) {
	ctx := context.Background()

	rpcClient, err := rpc.DialOptions(ctx, cfg.RPCURL,
		rpc.WithHTTPAuth(func(h http.Header) error {
			h.Set("Authorization", "Bearer "+cfg.APIToken)
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("beryx: failed to connect to RPC: %w", err)
	}

	ethClient := ethclient.NewClient(rpcClient)

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		rpcClient.Close()
		return nil, fmt.Errorf("beryx: failed to get chain ID: %w", err)
	}

	return &BeryxClient{
		ethClient: ethClient,
		rpcClient: rpcClient,
		chainID:   chainID,
		ctx:       ctx,
	}, nil
}

// ChainID returns the connected Filecoin network's chain ID.
// 314 = Mainnet, 314159 = Calibration.
func (c *BeryxClient) ChainID() *big.Int {
	return c.chainID
}

// Network returns the go-synapse network identifier based on chain ID.
func (c *BeryxClient) Network() constants.Network {
	if c.chainID.Int64() == int64(constants.ChainIDCalibration) {
		return constants.NetworkCalibration
	}
	return constants.NetworkMainnet
}

// BlockNumber returns the latest block height.
func (c *BeryxClient) BlockNumber() (uint64, error) {
	return c.ethClient.BlockNumber(c.ctx)
}

// EthClient returns the underlying go-ethereum client for direct use.
func (c *BeryxClient) EthClient() *ethclient.Client {
	return c.ethClient
}

// ListProviders queries the on-chain SP Registry for active storage providers.
func (c *BeryxClient) ListProviders() ([]*spregistry.ProviderInfo, error) {
	network := c.Network()
	registryAddr := constants.SPRegistryAddresses[network]

	registry, err := spregistry.NewService(c.ethClient, registryAddr, nil, c.chainID)
	if err != nil {
		return nil, fmt.Errorf("beryx: failed to create registry service: %w", err)
	}

	return registry.GetAllActiveProviders(c.ctx)
}

// Upload is not directly supported via Beryx RPC (read-only chain access).
// Returns an error directing the caller to use IPFS or a storage provider.
func (c *BeryxClient) Upload(report []byte) (string, error) {
	return "", fmt.Errorf("beryx: upload not supported via RPC — use IPFS or a Filecoin storage provider")
}

// Retrieve is not directly supported via Beryx RPC.
func (c *BeryxClient) Retrieve(cid string) ([]byte, error) {
	return nil, fmt.Errorf("beryx: retrieve not supported via RPC — use IPFS or a Filecoin gateway")
}

// Close releases the Beryx client resources.
func (c *BeryxClient) Close() {
	c.rpcClient.Close()
}

// --- Lighthouse Client (stable Filecoin/IPFS pinning) ---

// LighthouseClient stores data on Filecoin + IPFS via the Lighthouse.storage
// pinning service. Data is pinned to IPFS immediately and backed by Filecoin
// deals for long-term persistence. This is the primary storage backend.
type LighthouseClient struct {
	UploadURL  string
	GatewayURL string
	APIKey     string
	Client     *http.Client
}

// LighthouseConfig holds the configuration for Lighthouse.storage.
type LighthouseConfig struct {
	// APIKey is the Lighthouse API key for authentication.
	APIKey string
	// UploadURL is the upload endpoint. Default: https://upload.lighthouse.storage
	UploadURL string
	// GatewayURL is the IPFS gateway for retrieval. Default: https://gateway.lighthouse.storage
	GatewayURL string
}

// NewLighthouseClient creates a Lighthouse.storage client.
func NewLighthouseClient(cfg LighthouseConfig) *LighthouseClient {
	if cfg.UploadURL == "" {
		cfg.UploadURL = "https://upload.lighthouse.storage"
	}
	if cfg.GatewayURL == "" {
		cfg.GatewayURL = "https://gateway.lighthouse.storage"
	}
	return &LighthouseClient{
		UploadURL:  cfg.UploadURL,
		GatewayURL: cfg.GatewayURL,
		APIKey:     cfg.APIKey,
		Client:     &http.Client{Timeout: 60 * time.Second},
	}
}

// Upload stores a JSON report on Filecoin/IPFS via Lighthouse.
// Uses the /api/v0/add endpoint with Bearer auth and multipart form.
// Returns the IPFS CID (content hash).
func (c *LighthouseClient) Upload(report []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "report.json")
	if err != nil {
		return "", fmt.Errorf("lighthouse: failed to create form file: %w", err)
	}
	if _, err := part.Write(report); err != nil {
		return "", fmt.Errorf("lighthouse: failed to write report: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.UploadURL+"/api/v0/add", &buf)
	if err != nil {
		return "", fmt.Errorf("lighthouse: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("lighthouse: upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("lighthouse: upload returned status %d: %s", resp.StatusCode, string(body))
	}

	// Lighthouse returns JSON with a "Hash" field containing the IPFS CID.
	var result struct {
		Hash string `json:"Hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("lighthouse: failed to decode response: %w", err)
	}
	if result.Hash == "" {
		return "", fmt.Errorf("lighthouse: empty CID in response")
	}
	return result.Hash, nil
}

// Retrieve fetches a report from the Lighthouse IPFS gateway by CID.
func (c *LighthouseClient) Retrieve(cidStr string) ([]byte, error) {
	url := fmt.Sprintf("%s/ipfs/%s", c.GatewayURL, cidStr)
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("lighthouse: retrieve failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lighthouse: retrieve returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lighthouse: failed to read response: %w", err)
	}
	return data, nil
}

// --- Filecoin Client (via go-synapse, fallback) ---

// FilecoinClient stores data on Filecoin using the go-synapse SDK.
type FilecoinClient struct {
	client *synapse.Client
	ctx    context.Context
}

// FilecoinConfig holds the configuration for connecting to Filecoin.
type FilecoinConfig struct {
	PrivateKeyHex string
	RPCURL        string
	ProviderURL   string
}

// NewFilecoinClient creates a Filecoin storage client using go-synapse.
func NewFilecoinClient(cfg FilecoinConfig) (*FilecoinClient, error) {
	ctx := context.Background()

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("filecoin: invalid private key: %w", err)
	}

	client, err := synapse.New(ctx, synapse.Options{
		PrivateKey:  privateKey,
		RPCURL:      cfg.RPCURL,
		ProviderURL: cfg.ProviderURL,
	})
	if err != nil {
		return nil, fmt.Errorf("filecoin: failed to create client: %w", err)
	}

	return &FilecoinClient{client: client, ctx: ctx}, nil
}

// Upload stores a JSON report on Filecoin via the storage provider.
func (c *FilecoinClient) Upload(report []byte) (string, error) {
	storage, err := c.client.Storage()
	if err != nil {
		return "", fmt.Errorf("filecoin: failed to get storage manager: %w", err)
	}

	result, err := storage.UploadBytes(c.ctx, report, nil)
	if err != nil {
		return "", fmt.Errorf("filecoin: upload failed: %w", err)
	}

	cidStr := result.PieceCID.String()
	if cidStr == "" {
		return "", fmt.Errorf("filecoin: empty CID after upload")
	}
	return cidStr, nil
}

// Retrieve fetches a report from Filecoin by its piece CID.
func (c *FilecoinClient) Retrieve(cidStr string) ([]byte, error) {
	storage, err := c.client.Storage()
	if err != nil {
		return nil, fmt.Errorf("filecoin: failed to get storage manager: %w", err)
	}

	pieceCID, err := gocid.Decode(cidStr)
	if err != nil {
		return nil, fmt.Errorf("filecoin: invalid CID %q: %w", cidStr, err)
	}

	data, err := storage.Download(c.ctx, pieceCID, nil)
	if err != nil {
		return nil, fmt.Errorf("filecoin: download failed: %w", err)
	}
	return data, nil
}

// Close releases the Filecoin client resources.
func (c *FilecoinClient) Close() {
	c.client.Close()
}

// --- External IPFS Node Client ---

// IPFSClient communicates with a local IPFS node via its HTTP API.
type IPFSClient struct {
	APIURL string
	Client *http.Client
}

// NewIPFSClient creates a client for a local IPFS node.
func NewIPFSClient(apiURL string) *IPFSClient {
	return &IPFSClient{
		APIURL: apiURL,
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Upload adds a JSON report to the local IPFS node and returns the CID.
func (c *IPFSClient) Upload(report []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "report.json")
	if err != nil {
		return "", fmt.Errorf("ipfs: failed to create form file: %w", err)
	}
	if _, err := part.Write(report); err != nil {
		return "", fmt.Errorf("ipfs: failed to write report: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.APIURL+"/api/v0/add", &buf)
	if err != nil {
		return "", fmt.Errorf("ipfs: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ipfs: upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ipfs: upload returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Hash string `json:"Hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ipfs: failed to decode response: %w", err)
	}
	if result.Hash == "" {
		return "", fmt.Errorf("ipfs: empty CID in response")
	}
	return result.Hash, nil
}

// Retrieve fetches a report from the local IPFS node by CID.
func (c *IPFSClient) Retrieve(cid string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v0/cat?arg=%s", c.APIURL, cid)
	resp, err := c.Client.Post(url, "", nil)
	if err != nil {
		return nil, fmt.Errorf("ipfs: retrieve failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ipfs: retrieve returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ipfs: failed to read response: %w", err)
	}
	return data, nil
}

// --- Factory ---

// NewStore creates a Store based on environment configuration.
// Priority: Lighthouse (if API key set) > go-synapse Filecoin (if private key set) > IPFS (default).
// Use NewBeryxClient separately for chain queries.
func NewStore(ipfsURL string, lighthouseCfg *LighthouseConfig, filecoinCfg *FilecoinConfig) (Store, error) {
	if lighthouseCfg != nil && lighthouseCfg.APIKey != "" {
		return NewLighthouseClient(*lighthouseCfg), nil
	}
	if filecoinCfg != nil && filecoinCfg.PrivateKeyHex != "" {
		return NewFilecoinClient(*filecoinCfg)
	}
	if ipfsURL == "" {
		ipfsURL = "http://localhost:5001"
	}
	return NewIPFSClient(ipfsURL), nil
}
