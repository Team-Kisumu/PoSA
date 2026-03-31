package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Murzuqisah/PoSA/blockchain"
)

// mockStore implements storage.Store for testing.
type mockStore struct {
	data     map[string][]byte
	failNext bool
}

func newMockStore() *mockStore {
	return &mockStore{data: make(map[string][]byte)}
}

func (m *mockStore) Upload(report []byte) (string, error) {
	if m.failNext {
		m.failNext = false
		return "", fmt.Errorf("mock: upload failed")
	}
	cid := fmt.Sprintf("QmMock%x", len(m.data))
	m.data[cid] = report
	return cid, nil
}

func (m *mockStore) Retrieve(cid string) ([]byte, error) {
	data, ok := m.data[cid]
	if !ok {
		return nil, fmt.Errorf("mock: not found")
	}
	return data, nil
}

// newMockFlowServer creates a test HTTP server that mimics the Flow REST API.
func newMockFlowServer(t *testing.T, failCount int) *httptest.Server {
	calls := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/v1/transactions" {
			if calls <= failCount {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("temporary error"))
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"id": "tx-mock-123"})
			return
		}
		if r.URL.Path == "/v1/scripts" {
			json.NewEncoder(w).Encode(blockchain.Credential{
				ID: 1, CID: "QmMock0", Score: 85, Submitter: "0xTest",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// TestStoreAndAnchorSuccess tests the full pipeline: store + anchor.
func TestStoreAndAnchorSuccess(t *testing.T) {
	store := newMockStore()
	server := newMockFlowServer(t, 0)
	defer server.Close()

	chain := blockchain.NewFlowClient(blockchain.FlowConfig{
		AccessNode: server.URL, ContractAddress: "0xTest",
	})
	pipe := NewPipeline(store, chain)

	report := &EvaluationReport{
		Score:       85,
		Issues:      []any{},
		Suggestions: []string{"looks good"},
		Filename:    "main.go",
	}

	result, err := pipe.StoreAndAnchor(report, "0xSubmitter")
	if err != nil {
		t.Fatalf("StoreAndAnchor failed: %v", err)
	}
	if result.CID == "" {
		t.Error("empty CID")
	}
	if result.TxHash != "tx-mock-123" {
		t.Errorf("TxHash = %q, want tx-mock-123", result.TxHash)
	}
	if result.Score != 85 {
		t.Errorf("Score = %d, want 85", result.Score)
	}
	t.Logf("Pipeline result: CID=%s TxHash=%s", result.CID, result.TxHash)
}

// TestStoreAndAnchorStorageFail tests pipeline when storage fails.
func TestStoreAndAnchorStorageFail(t *testing.T) {
	store := newMockStore()
	store.failNext = true
	server := newMockFlowServer(t, 0)
	defer server.Close()

	chain := blockchain.NewFlowClient(blockchain.FlowConfig{
		AccessNode: server.URL, ContractAddress: "0xTest",
	})
	pipe := NewPipeline(store, chain)

	_, err := pipe.StoreAndAnchor(&EvaluationReport{Score: 85, Filename: "test.go"}, "0xSub")
	if err == nil {
		t.Error("expected error when storage fails")
	}
}

// TestStoreAndAnchorBlockchainFail tests pipeline when blockchain fails after storage succeeds.
func TestStoreAndAnchorBlockchainFail(t *testing.T) {
	store := newMockStore()
	server := newMockFlowServer(t, 100) // Always fail
	defer server.Close()

	chain := blockchain.NewFlowClient(blockchain.FlowConfig{
		AccessNode: server.URL, ContractAddress: "0xTest",
	})
	pipe := NewPipeline(store, chain)

	result, err := pipe.StoreAndAnchor(&EvaluationReport{Score: 90, Filename: "test.py"}, "0xSub")
	if err == nil {
		t.Error("expected error when blockchain fails")
	}
	// CID should still be returned even if anchor fails.
	if result == nil || result.CID == "" {
		t.Error("expected CID even when anchor fails")
	}
	t.Logf("Partial result: CID=%s (anchor failed as expected)", result.CID)
}

// TestStoreAndAnchorRetry tests that retry logic recovers from transient failures.
func TestStoreAndAnchorRetry(t *testing.T) {
	store := newMockStore()
	server := newMockFlowServer(t, 2) // Fail first 2 attempts, succeed on 3rd
	defer server.Close()

	chain := blockchain.NewFlowClient(blockchain.FlowConfig{
		AccessNode: server.URL, ContractAddress: "0xTest",
	})
	pipe := NewPipeline(store, chain)

	result, err := pipe.StoreAndAnchor(&EvaluationReport{Score: 75, Filename: "retry.js"}, "0xSub")
	if err != nil {
		t.Fatalf("StoreAndAnchor failed after retry: %v", err)
	}
	if result.TxHash != "tx-mock-123" {
		t.Errorf("TxHash = %q, want tx-mock-123", result.TxHash)
	}
	t.Logf("Retry succeeded: CID=%s TxHash=%s", result.CID, result.TxHash)
}

// TestVerifySuccess tests the verify pipeline.
func TestVerifySuccess(t *testing.T) {
	store := newMockStore()
	store.data["QmMock0"] = []byte(`{"score":85}`)
	server := newMockFlowServer(t, 0)
	defer server.Close()

	chain := blockchain.NewFlowClient(blockchain.FlowConfig{
		AccessNode: server.URL, ContractAddress: "0xTest",
	})
	pipe := NewPipeline(store, chain)

	cred, data, err := pipe.Verify("QmMock0")
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if cred == nil {
		t.Fatal("expected credential")
	}
	if cred.Score != 85 {
		t.Errorf("Score = %d, want 85", cred.Score)
	}
	if string(data) != `{"score":85}` {
		t.Errorf("data = %q", string(data))
	}
}

// TestTimestampAutoSet tests that timestamp is set automatically.
func TestTimestampAutoSet(t *testing.T) {
	store := newMockStore()
	server := newMockFlowServer(t, 0)
	defer server.Close()

	chain := blockchain.NewFlowClient(blockchain.FlowConfig{
		AccessNode: server.URL, ContractAddress: "0xTest",
	})
	pipe := NewPipeline(store, chain)

	report := &EvaluationReport{Score: 50, Filename: "test.go"}
	_, err := pipe.StoreAndAnchor(report, "0xSub")
	if err != nil {
		t.Fatalf("StoreAndAnchor failed: %v", err)
	}
	if report.Timestamp == "" {
		t.Error("expected timestamp to be auto-set")
	}
}
