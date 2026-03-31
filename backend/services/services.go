// Package services orchestrates the PoSA evaluation pipeline.
// It connects the storage layer (Lighthouse/IPFS) with the blockchain
// layer (Flow) to store evaluation reports and anchor proofs on-chain.
package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Murzuqisah/PoSA/blockchain"
	"github.com/Murzuqisah/PoSA/storage"
)

// EvaluationReport is the JSON report stored on IPFS/Filecoin.
type EvaluationReport struct {
	Score       int      `json:"score"`
	Issues      []any    `json:"issues"`
	Suggestions []string `json:"suggestions"`
	Filename    string   `json:"filename"`
	Language    string   `json:"language,omitempty"`
	Timestamp   string   `json:"timestamp"`
}

// AnchorResult is returned after storing a report and anchoring it on-chain.
type AnchorResult struct {
	CID       string `json:"cid"`
	TxHash    string `json:"tx_hash"`
	Score     int    `json:"score"`
	Submitter string `json:"submitter"`
}

// Pipeline orchestrates the store → anchor flow.
type Pipeline struct {
	store storage.Store
	chain *blockchain.FlowClient
}

// NewPipeline creates a new pipeline with the given storage and blockchain clients.
func NewPipeline(store storage.Store, chain *blockchain.FlowClient) *Pipeline {
	return &Pipeline{store: store, chain: chain}
}

// StoreAndAnchor stores an evaluation report on IPFS/Filecoin and anchors
// the CID on the Flow blockchain. Returns the CID and transaction hash.
func (p *Pipeline) StoreAndAnchor(report *EvaluationReport, submitter string) (*AnchorResult, error) {
	// Set timestamp if not already set.
	if report.Timestamp == "" {
		report.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// Serialize the report to JSON.
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("pipeline: failed to serialize report: %w", err)
	}

	// Step 1: Upload to storage (Lighthouse/IPFS).
	cid, err := p.store.Upload(reportJSON)
	if err != nil {
		return nil, fmt.Errorf("pipeline: storage upload failed: %w", err)
	}

	// Step 2: Anchor on blockchain with retry.
	txHash, err := p.anchorWithRetry(cid, uint8(report.Score), submitter, 3)
	if err != nil {
		// Storage succeeded but blockchain failed — return CID with error context.
		return &AnchorResult{
			CID:       cid,
			Score:     report.Score,
			Submitter: submitter,
		}, fmt.Errorf("pipeline: stored (CID=%s) but anchor failed: %w", cid, err)
	}

	return &AnchorResult{
		CID:       cid,
		TxHash:    txHash,
		Score:     report.Score,
		Submitter: submitter,
	}, nil
}

// Verify checks if a credential exists on-chain for the given CID
// and optionally retrieves the report from storage.
func (p *Pipeline) Verify(cid string) (*blockchain.Credential, []byte, error) {
	// Query blockchain for the credential.
	credential, err := p.chain.VerifyProof(cid)
	if err != nil {
		return nil, nil, fmt.Errorf("pipeline: blockchain verify failed: %w", err)
	}
	if credential == nil {
		return nil, nil, nil
	}

	// Retrieve the report from storage.
	data, err := p.store.Retrieve(cid)
	if err != nil {
		// Credential exists on-chain but report retrieval failed.
		return credential, nil, fmt.Errorf("pipeline: credential found but report retrieval failed: %w", err)
	}

	return credential, data, nil
}

// anchorWithRetry attempts to anchor a proof on-chain with exponential backoff.
func (p *Pipeline) anchorWithRetry(cid string, score uint8, submitter string, maxRetries int) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		txHash, err := p.chain.AnchorProof(cid, score, submitter)
		if err == nil {
			return txHash, nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("failed after %d retries: %w", maxRetries+1, lastErr)
}
