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
		// Verify multipart form contains the report.
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
		// Return mock CID.
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

// --- Filecoin Client Tests (via web3.storage) ---

// TestFilecoinUpload verifies that the Filecoin client uploads a report
// with the correct authorization header and returns the CID.
func TestFilecoinUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify auth header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-w3s-token" {
			t.Errorf("auth = %q, want Bearer test-w3s-token", auth)
		}
		// Verify multipart form.
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("failed to parse multipart: %v", err)
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("missing file field: %v", err)
		}
		data, _ := io.ReadAll(file)
		if string(data) != `{"score":92}` {
			t.Errorf("unexpected body: %s", string(data))
		}
		json.NewEncoder(w).Encode(map[string]string{
			"cid": "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
		})
	}))
	defer server.Close()

	client := NewFilecoinClient(server.URL, "", "test-w3s-token")
	cid, err := client.Upload([]byte(`{"score":92}`))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if cid != "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi" {
		t.Errorf("cid = %q", cid)
	}
}

// TestFilecoinUploadError verifies error handling for Filecoin upload failures.
func TestFilecoinUploadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid token"))
	}))
	defer server.Close()

	client := NewFilecoinClient(server.URL, "", "bad-token")
	_, err := client.Upload([]byte(`{"score":92}`))
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

// TestFilecoinUploadEmptyCID verifies error handling when web3.storage returns an empty CID.
func TestFilecoinUploadEmptyCID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"cid": ""})
	}))
	defer server.Close()

	client := NewFilecoinClient(server.URL, "", "token")
	_, err := client.Upload([]byte(`{"score":92}`))
	if err == nil {
		t.Error("expected error for empty CID")
	}
}

// TestFilecoinRetrieve verifies that the Filecoin client fetches via the gateway.
func TestFilecoinRetrieve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ipfs/bafyTestCID" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"score":92}`))
	}))
	defer server.Close()

	client := NewFilecoinClient("", server.URL, "token")
	data, err := client.Retrieve("bafyTestCID")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if string(data) != `{"score":92}` {
		t.Errorf("data = %q", string(data))
	}
}

// TestFilecoinRetrieveError verifies error handling when gateway returns 404.
func TestFilecoinRetrieveError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := NewFilecoinClient("", server.URL, "token")
	_, err := client.Retrieve("bafyNonExistent")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

// --- Factory Tests ---

// TestNewStoreIPFS verifies the factory returns an IPFS client when no Filecoin token is set.
func TestNewStoreIPFS(t *testing.T) {
	store := NewStore("http://localhost:5001", "", "", "")
	if _, ok := store.(*IPFSClient); !ok {
		t.Error("expected IPFSClient when no Filecoin token")
	}
}

// TestNewStoreFilecoin verifies the factory returns a Filecoin client when a token is provided.
func TestNewStoreFilecoin(t *testing.T) {
	store := NewStore("", "https://api.web3.storage", "", "my-w3s-token")
	if _, ok := store.(*FilecoinClient); !ok {
		t.Error("expected FilecoinClient when token is set")
	}
}

// TestNewStoreDefaultIPFS verifies the factory uses default IPFS URL when none is provided.
func TestNewStoreDefaultIPFS(t *testing.T) {
	store := NewStore("", "", "", "")
	client, ok := store.(*IPFSClient)
	if !ok {
		t.Fatal("expected IPFSClient")
	}
	if client.APIURL != "http://localhost:5001" {
		t.Errorf("APIURL = %q, want http://localhost:5001", client.APIURL)
	}
}

// TestNewFilecoinClientDefaults verifies default URLs for the Filecoin client.
func TestNewFilecoinClientDefaults(t *testing.T) {
	client := NewFilecoinClient("", "", "token")
	if client.APIURL != "https://api.web3.storage" {
		t.Errorf("APIURL = %q, want https://api.web3.storage", client.APIURL)
	}
	if client.GatewayURL != "https://w3s.link" {
		t.Errorf("GatewayURL = %q, want https://w3s.link", client.GatewayURL)
	}
}
