// Package blockchain provides the integration layer between the Go backend
// and the Flow blockchain. It handles submitting transactions to mint
// proof credentials and querying the contract to verify them.
package blockchain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Credential represents a proof-of-skill credential stored on-chain.
type Credential struct {
	ID        uint64  `json:"id"`
	CID       string  `json:"cid"`
	Score     uint8   `json:"score"`
	Submitter string  `json:"submitter"`
	Timestamp float64 `json:"timestamp"`
}

// FlowConfig holds the configuration for connecting to the Flow network.
type FlowConfig struct {
	// AccessNode is the Flow access node URL.
	// Testnet: https://rest-testnet.onflow.org
	// Mainnet: https://rest-mainnet.onflow.org
	AccessNode string
	// AccountAddress is the Flow account that deployed the contract.
	AccountAddress string
	// PrivateKey is the account's private key for signing transactions.
	PrivateKey string
	// ContractAddress is the address where ProofOfSkill is deployed.
	ContractAddress string
}

// FlowClient interacts with the ProofOfSkill contract on Flow.
type FlowClient struct {
	config FlowConfig
	client *http.Client
}

// NewFlowClient creates a Flow blockchain client.
func NewFlowClient(cfg FlowConfig) *FlowClient {
	return &FlowClient{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// AnchorProof mints a new proof credential on-chain.
// Returns the transaction ID on success.
func (c *FlowClient) AnchorProof(cid string, score uint8, submitter string) (string, error) {
	if cid == "" {
		return "", fmt.Errorf("flow: CID must not be empty")
	}
	if score > 100 {
		return "", fmt.Errorf("flow: score must be 0-100, got %d", score)
	}

	// Build the Cadence transaction script.
	script := fmt.Sprintf(`
		import ProofOfSkill from %s

		transaction {
			let admin: &ProofOfSkill.Admin

			prepare(signer: auth(BorrowValue) &Account) {
				self.admin = signer.storage.borrow<&ProofOfSkill.Admin>(
					from: /storage/ProofOfSkillAdmin
				) ?? panic("Admin resource not found")
			}

			execute {
				self.admin.mintCredential(
					cid: "%s",
					score: %d,
					submitter: %s
				)
			}
		}
	`, c.config.ContractAddress, cid, score, submitter)

	// Submit transaction via Flow REST API.
	txID, err := c.submitTransaction(script)
	if err != nil {
		return "", fmt.Errorf("flow: anchor failed: %w", err)
	}
	return txID, nil
}

// VerifyProof queries the contract for a credential by CID.
// Returns nil if no credential exists.
func (c *FlowClient) VerifyProof(cid string) (*Credential, error) {
	if cid == "" {
		return nil, fmt.Errorf("flow: CID must not be empty")
	}

	script := fmt.Sprintf(`
		import ProofOfSkill from %s

		access(all) fun main(cid: String): ProofOfSkill.Credential? {
			return ProofOfSkill.getCredential(cid: cid)
		}
	`, c.config.ContractAddress)

	result, err := c.executeScript(script, cid)
	if err != nil {
		return nil, fmt.Errorf("flow: verify failed: %w", err)
	}
	if result == nil {
		return nil, nil
	}
	return result, nil
}

// HasProof checks if a credential exists for the given CID.
func (c *FlowClient) HasProof(cid string) (bool, error) {
	credential, err := c.VerifyProof(cid)
	if err != nil {
		return false, err
	}
	return credential != nil, nil
}

// submitTransaction sends a transaction to the Flow network.
// Returns the transaction ID. This is a simplified implementation —
// production use should use the Flow Go SDK for proper key signing.
func (c *FlowClient) submitTransaction(script string) (string, error) {
	payload := map[string]any{
		"script":    script,
		"arguments": []any{},
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/v1/transactions", c.config.AccessNode)
	resp, err := c.client.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	json.Unmarshal(respBody, &result)
	return result.ID, nil
}

// executeScript runs a read-only Cadence script on the Flow network.
func (c *FlowClient) executeScript(script string, args ...string) (*Credential, error) {
	payload := map[string]any{
		"script":    script,
		"arguments": args,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/v1/scripts", c.config.AccessNode)
	resp, err := c.client.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the Cadence JSON response into a Credential.
	var credential Credential
	if err := json.Unmarshal(respBody, &credential); err != nil {
		return nil, nil // No credential found or unparseable response.
	}
	if credential.CID == "" {
		return nil, nil
	}
	return &credential, nil
}
