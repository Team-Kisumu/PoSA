package storage

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- IPFS Client Tests ---

// TestIPFSUpload verifies that the IPFS client uploads a report
// and returns the CID from the /api/v0/add response.
func TestIPFSUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/add" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("failed to parse multipart: %v", err)
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("missing file field: %v", err)
		}
		data, _ := io.ReadAll(file)
		if string(data) != `{"score":85}` {
			t.Errorf("unexpected body: %s", string(data))
		}
		json.NewEncoder(w).Encode(map[string]string{"Hash": "QmTestCID123456789012345678901234567890abcdef"})
	}))
	defer server.Close()

	client := NewIPFSClient(server.URL)
	cid, err := client.Upload([]byte(`{"score":85}`))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if cid != "QmTestCID123456789012345678901234567890abcdef" {
		t.Errorf("cid = %q, want QmTestCID...", cid)
	}
}

// TestIPFSUploadError verifies error handling when the IPFS node returns an error.
func TestIPFSUploadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("node unavailable"))
	}))
	defer server.Close()

	client := NewIPFSClient(server.URL)
	_, err := client.Upload([]byte(`{"score":85}`))
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

// TestIPFSUploadEmptyCID verifies error handling when IPFS returns an empty hash.
func TestIPFSUploadEmptyCID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"Hash": ""})
	}))
	defer server.Close()

	client := NewIPFSClient(server.URL)
	_, err := client.Upload([]byte(`{"score":85}`))
	if err == nil {
		t.Error("expected error for empty CID")
	}
}

// TestIPFSRetrieve verifies that the IPFS client fetches a report by CID.
func TestIPFSRetrieve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/cat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		cid := r.URL.Query().Get("arg")
		if cid != "QmTestCID" {
			t.Errorf("cid = %q, want QmTestCID", cid)
		}
		w.Write([]byte(`{"score":85,"issues":[]}`))
	}))
	defer server.Close()

	client := NewIPFSClient(server.URL)
	data, err := client.Retrieve("QmTestCID")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if string(data) != `{"score":85,"issues":[]}` {
		t.Errorf("data = %q", string(data))
	}
}

// TestIPFSRetrieveError verifies error handling when retrieval fails.
func TestIPFSRetrieveError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := NewIPFSClient(server.URL)
	_, err := client.Retrieve("QmNonExistent")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

// --- Factory Tests ---

// TestNewStoreIPFS verifies the factory returns an IPFS client when no Filecoin config is set.
func TestNewStoreIPFS(t *testing.T) {
	store, err := NewStore("http://localhost:5001", nil)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if _, ok := store.(*IPFSClient); !ok {
		t.Error("expected IPFSClient when no Filecoin config")
	}
}

// TestNewStoreDefaultIPFS verifies the factory uses default IPFS URL when none is provided.
func TestNewStoreDefaultIPFS(t *testing.T) {
	store, err := NewStore("", nil)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	client, ok := store.(*IPFSClient)
	if !ok {
		t.Fatal("expected IPFSClient")
	}
	if client.APIURL != "http://localhost:5001" {
		t.Errorf("APIURL = %q, want http://localhost:5001", client.APIURL)
	}
}

// TestFilecoinConfigValidation verifies that invalid private keys are rejected.
func TestFilecoinConfigValidation(t *testing.T) {
	_, err := NewFilecoinClient(FilecoinConfig{
		PrivateKeyHex: "not-a-valid-hex-key",
		RPCURL:        "https://api.calibration.node.glif.io/rpc/v1",
		ProviderURL:   "https://provider.example.com",
	})
	if err == nil {
		t.Error("expected error for invalid private key")
	}
}
