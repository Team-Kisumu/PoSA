// Package storage provides content-addressed storage for evaluation reports.
// It supports two backends: a local IPFS node (HTTP API) and Filecoin
// via the web3.storage API (which pins to IPFS and creates Filecoin deals).
// Both implement the Store interface for swappable use.
package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Store defines the interface for content-addressed storage backends.
// Upload stores a JSON report and returns its CID.
// Retrieve fetches a report by its CID.
type Store interface {
	Upload(report []byte) (string, error)
	Retrieve(cid string) ([]byte, error)
}

// --- Local IPFS Node Client ---

// IPFSClient communicates with a local IPFS node via its HTTP API.
// Default endpoint: http://localhost:5001/api/v0
type IPFSClient struct {
	APIURL string
	Client *http.Client
}

// NewIPFSClient creates a client for a local IPFS node.
// apiURL should be the base URL of the IPFS HTTP API (e.g. http://localhost:5001).
func NewIPFSClient(apiURL string) *IPFSClient {
	return &IPFSClient{
		APIURL: apiURL,
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Upload adds a JSON report to the local IPFS node and returns the CID.
// Uses the /api/v0/add endpoint with multipart form upload.
func (c *IPFSClient) Upload(report []byte) (string, error) {
	// Build multipart form body with the report as a file field.
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

	// IPFS /api/v0/add returns JSON with a "Hash" field containing the CID.
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
// Uses the /api/v0/cat endpoint.
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

// --- Filecoin Client (via web3.storage) ---

// FilecoinClient stores data on IPFS + Filecoin via the web3.storage API.
// Data is pinned to IPFS immediately and backed by Filecoin storage deals
// for long-term persistence. Requires a web3.storage API token.
type FilecoinClient struct {
	APIURL     string
	GatewayURL string
	APIToken   string
	Client     *http.Client
}

// NewFilecoinClient creates a client for Filecoin storage via web3.storage.
// apiToken should be a web3.storage API token.
func NewFilecoinClient(apiURL, gatewayURL, apiToken string) *FilecoinClient {
	if apiURL == "" {
		apiURL = "https://api.web3.storage"
	}
	if gatewayURL == "" {
		gatewayURL = "https://w3s.link"
	}
	return &FilecoinClient{
		APIURL:     apiURL,
		GatewayURL: gatewayURL,
		APIToken:   apiToken,
		Client:     &http.Client{Timeout: 60 * time.Second},
	}
}

// Upload stores a JSON report on IPFS + Filecoin via web3.storage.
// Uses the /upload endpoint. Data is pinned to IPFS immediately and
// a Filecoin storage deal is created asynchronously for persistence.
func (c *FilecoinClient) Upload(report []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "report.json")
	if err != nil {
		return "", fmt.Errorf("filecoin: failed to create form file: %w", err)
	}
	if _, err := part.Write(report); err != nil {
		return "", fmt.Errorf("filecoin: failed to write report: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.APIURL+"/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("filecoin: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("filecoin: upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("filecoin: upload returned status %d: %s", resp.StatusCode, string(body))
	}

	// web3.storage returns JSON with a "cid" field.
	var result struct {
		CID string `json:"cid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("filecoin: failed to decode response: %w", err)
	}
	if result.CID == "" {
		return "", fmt.Errorf("filecoin: empty CID in response")
	}
	return result.CID, nil
}

// Retrieve fetches a report from the IPFS/Filecoin gateway by CID.
// Uses the w3s.link gateway (or configured gateway URL).
func (c *FilecoinClient) Retrieve(cid string) ([]byte, error) {
	url := fmt.Sprintf("%s/ipfs/%s", c.GatewayURL, cid)
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("filecoin: retrieve failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("filecoin: retrieve returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("filecoin: failed to read response: %w", err)
	}
	return data, nil
}

// --- Factory ---

// NewStore creates a Store based on environment configuration.
// If filecoinToken is non-empty, uses Filecoin (via web3.storage).
// Otherwise uses a local IPFS node.
func NewStore(ipfsURL, filecoinAPIURL, filecoinGatewayURL, filecoinToken string) Store {
	if filecoinToken != "" {
		return NewFilecoinClient(filecoinAPIURL, filecoinGatewayURL, filecoinToken)
	}
	if ipfsURL == "" {
		ipfsURL = "http://localhost:5001"
	}
	return NewIPFSClient(ipfsURL)
}
