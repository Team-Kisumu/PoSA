// Package storage provides content-addressed storage for evaluation reports.
// It supports two backends: a local IPFS node (HTTP API) and the Pinata
// cloud pinning service. Both implement the Store interface for swappable use.
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

// --- Pinata Cloud Client ---

// PinataClient communicates with the Pinata IPFS pinning service.
// Requires a JWT API token for authentication.
type PinataClient struct {
	APIURL string
	APIKey string
	Client *http.Client
}

// NewPinataClient creates a client for the Pinata pinning service.
// apiKey should be a Pinata JWT token.
func NewPinataClient(apiURL, apiKey string) *PinataClient {
	if apiURL == "" {
		apiURL = "https://api.pinata.cloud"
	}
	return &PinataClient{
		APIURL: apiURL,
		APIKey: apiKey,
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Upload pins a JSON report to Pinata and returns the CID.
// Uses the /pinning/pinFileToIPFS endpoint.
func (c *PinataClient) Upload(report []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "report.json")
	if err != nil {
		return "", fmt.Errorf("pinata: failed to create form file: %w", err)
	}
	if _, err := part.Write(report); err != nil {
		return "", fmt.Errorf("pinata: failed to write report: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.APIURL+"/pinning/pinFileToIPFS", &buf)
	if err != nil {
		return "", fmt.Errorf("pinata: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("pinata: upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("pinata: upload returned status %d: %s", resp.StatusCode, string(body))
	}

	// Pinata returns JSON with "IpfsHash" field.
	var result struct {
		IpfsHash string `json:"IpfsHash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("pinata: failed to decode response: %w", err)
	}
	if result.IpfsHash == "" {
		return "", fmt.Errorf("pinata: empty CID in response")
	}
	return result.IpfsHash, nil
}

// Retrieve fetches a report from the Pinata gateway by CID.
// Uses the public IPFS gateway at gateway.pinata.cloud.
func (c *PinataClient) Retrieve(cid string) ([]byte, error) {
	url := fmt.Sprintf("https://gateway.pinata.cloud/ipfs/%s", cid)
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("pinata: retrieve failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pinata: retrieve returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("pinata: failed to read response: %w", err)
	}
	return data, nil
}

// --- Factory ---

// NewStore creates a Store based on environment configuration.
// If pinataKey is non-empty, uses Pinata. Otherwise uses local IPFS.
func NewStore(ipfsURL, pinataURL, pinataKey string) Store {
	if pinataKey != "" {
		return NewPinataClient(pinataURL, pinataKey)
	}
	if ipfsURL == "" {
		ipfsURL = "http://localhost:5001"
	}
	return NewIPFSClient(ipfsURL)
}
