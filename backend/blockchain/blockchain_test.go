package blockchain

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAnchorProofValidation verifies input validation before submitting.
func TestAnchorProofValidation(t *testing.T) {
	client := NewFlowClient(FlowConfig{AccessNode: "http://localhost"})

	t.Run("empty CID rejected", func(t *testing.T) {
		_, err := client.AnchorProof("", 85, "0x1234")
		if err == nil {
			t.Error("expected error for empty CID")
		}
	})

	t.Run("score above 100 rejected", func(t *testing.T) {
		_, err := client.AnchorProof("QmTest", 101, "0x1234")
		if err == nil {
			t.Error("expected error for score > 100")
		}
	})
}

// TestVerifyProofValidation verifies input validation for verify.
func TestVerifyProofValidation(t *testing.T) {
	client := NewFlowClient(FlowConfig{AccessNode: "http://localhost"})

	_, err := client.VerifyProof("")
	if err == nil {
		t.Error("expected error for empty CID")
	}
}

// TestAnchorProofSuccess verifies a successful transaction submission.
func TestAnchorProofSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transactions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "tx-abc123"})
	}))
	defer server.Close()

	client := NewFlowClient(FlowConfig{
		AccessNode:      server.URL,
		ContractAddress: "0xPROOF",
	})

	txID, err := client.AnchorProof("QmTestCID", 85, "0xSubmitter")
	if err != nil {
		t.Fatalf("AnchorProof failed: %v", err)
	}
	if txID != "tx-abc123" {
		t.Errorf("txID = %q, want tx-abc123", txID)
	}
}

// TestAnchorProofServerError verifies error handling for failed transactions.
func TestAnchorProofServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewFlowClient(FlowConfig{AccessNode: server.URL, ContractAddress: "0xPROOF"})

	_, err := client.AnchorProof("QmTestCID", 85, "0xSubmitter")
	if err == nil {
		t.Error("expected error for server error")
	}
}

// TestVerifyProofFound verifies successful credential lookup.
func TestVerifyProofFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Credential{
			ID:        1,
			CID:       "QmTestCID",
			Score:     85,
			Submitter: "0xSubmitter",
			Timestamp: 1234567890.0,
		})
	}))
	defer server.Close()

	client := NewFlowClient(FlowConfig{AccessNode: server.URL, ContractAddress: "0xPROOF"})

	cred, err := client.VerifyProof("QmTestCID")
	if err != nil {
		t.Fatalf("VerifyProof failed: %v", err)
	}
	if cred == nil {
		t.Fatal("expected credential, got nil")
	}
	if cred.CID != "QmTestCID" {
		t.Errorf("CID = %q, want QmTestCID", cred.CID)
	}
	if cred.Score != 85 {
		t.Errorf("Score = %d, want 85", cred.Score)
	}
}

// TestVerifyProofNotFound verifies nil return for missing credential.
func TestVerifyProofNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("null"))
	}))
	defer server.Close()

	client := NewFlowClient(FlowConfig{AccessNode: server.URL, ContractAddress: "0xPROOF"})

	cred, err := client.VerifyProof("QmNonExistent")
	if err != nil {
		t.Fatalf("VerifyProof failed: %v", err)
	}
	if cred != nil {
		t.Errorf("expected nil, got %+v", cred)
	}
}

// TestHasProof verifies the convenience boolean check.
func TestHasProof(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Credential{CID: "QmExists", Score: 90})
	}))
	defer server.Close()

	client := NewFlowClient(FlowConfig{AccessNode: server.URL, ContractAddress: "0xPROOF"})

	has, err := client.HasProof("QmExists")
	if err != nil {
		t.Fatalf("HasProof failed: %v", err)
	}
	if !has {
		t.Error("expected true for existing credential")
	}
}

// TestNewFlowClient verifies client creation.
func TestNewFlowClient(t *testing.T) {
	client := NewFlowClient(FlowConfig{
		AccessNode:      "https://rest-testnet.onflow.org",
		AccountAddress:  "0x1234",
		ContractAddress: "0x5678",
	})
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
