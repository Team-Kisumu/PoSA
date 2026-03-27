// Package storage provides content-addressed storage for evaluation reports.
//
// Three backends are supported:
//   - FilecoinClient: Filecoin storage via go-synapse SDK (production)
//   - IPFSClient: External IPFS node via HTTP API (development)
//
// All implement the Store interface for swappable use.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/data-preservation-programs/go-synapse"
	"github.com/ethereum/go-ethereum/crypto"
	gocid "github.com/ipfs/go-cid"
)

// Store defines the interface for content-addressed storage backends.
// Upload stores a JSON report and returns its CID.
// Retrieve fetches a report by its CID.
type Store interface {
	Upload(report []byte) (string, error)
	Retrieve(cid string) ([]byte, error)
}

// --- Filecoin Client (via go-synapse) ---

// FilecoinClient stores data on Filecoin using the go-synapse SDK.
// It handles file upload to a storage provider and proof set management
// for on-chain data possession verification.
type FilecoinClient struct {
	client *synapse.Client
	ctx    context.Context
}

// FilecoinConfig holds the configuration for connecting to Filecoin.
type FilecoinConfig struct {
	// PrivateKeyHex is the hex-encoded ECDSA private key for signing transactions.
	PrivateKeyHex string
	// RPCURL is the Filecoin RPC endpoint (e.g. calibration testnet).
	RPCURL string
	// ProviderURL is the storage provider's API endpoint.
	ProviderURL string
}

// NewFilecoinClient creates a Filecoin storage client using go-synapse.
// Connects to the Filecoin network and authenticates with the storage provider.
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
// Returns the piece CID as a string.
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

	// Parse the CID string into a go-cid object.
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
// Default endpoint: http://localhost:5001/api/v0
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
// If Filecoin config is complete, uses go-synapse FilecoinClient.
// Otherwise falls back to local IPFS node.
func NewStore(ipfsURL string, filecoinCfg *FilecoinConfig) (Store, error) {
	if filecoinCfg != nil && filecoinCfg.PrivateKeyHex != "" {
		return NewFilecoinClient(*filecoinCfg)
	}
	if ipfsURL == "" {
		ipfsURL = "http://localhost:5001"
	}
	return NewIPFSClient(ipfsURL), nil
}
