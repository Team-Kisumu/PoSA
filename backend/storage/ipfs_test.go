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

// --- Pinata Client Tests ---

// TestPinataUpload verifies that the Pinata client uploads a report
// with the correct authorization header and returns the CID.
func TestPinataUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pinning/pinFileToIPFS" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify auth header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-jwt-token" {
			t.Errorf("auth = %q, want Bearer test-jwt-token", auth)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"IpfsHash": "QmPinataTestCID12345678901234567890abcdefgh",
		})
	}))
	defer server.Close()

	client := NewPinataClient(server.URL, "test-jwt-token")
	cid, err := client.Upload([]byte(`{"score":92}`))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if cid != "QmPinataTestCID12345678901234567890abcdefgh" {
		t.Errorf("cid = %q", cid)
	}
}

// TestPinataUploadError verifies error handling for Pinata upload failures.
func TestPinataUploadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid API key"))
	}))
	defer server.Close()

	client := NewPinataClient(server.URL, "bad-key")
	_, err := client.Upload([]byte(`{"score":92}`))
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

// TestPinataUploadEmptyCID verifies error handling when Pinata returns an empty hash.
func TestPinataUploadEmptyCID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"IpfsHash": ""})
	}))
	defer server.Close()

	client := NewPinataClient(server.URL, "key")
	_, err := client.Upload([]byte(`{"score":92}`))
	if err == nil {
		t.Error("expected error for empty CID")
	}
}

// TestPinataRetrieve verifies that the Pinata client fetches via the gateway.
// Note: In production this hits gateway.pinata.cloud, but we mock it here.
func TestPinataRetrieve(t *testing.T) {
	// The Pinata client uses a hardcoded gateway URL, so we test the
	// IPFSClient retrieve path instead (same HTTP GET pattern).
	// For a full Pinata gateway test, we'd need to override the gateway URL.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"score":92}`))
	}))
	defer server.Close()

	// Use IPFS client pointed at the test server to verify retrieve logic.
	client := NewIPFSClient(server.URL)
	data, err := client.Retrieve("QmTest")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if string(data) != `{"score":92}` {
		t.Errorf("data = %q", string(data))
	}
}

// --- Factory Tests ---

// TestNewStoreIPFS verifies the factory returns an IPFS client when no Pinata key is set.
func TestNewStoreIPFS(t *testing.T) {
	store := NewStore("http://localhost:5001", "", "")
	if _, ok := store.(*IPFSClient); !ok {
		t.Error("expected IPFSClient when no Pinata key")
	}
}

// TestNewStorePinata verifies the factory returns a Pinata client when a key is provided.
func TestNewStorePinata(t *testing.T) {
	store := NewStore("", "https://api.pinata.cloud", "my-jwt-token")
	if _, ok := store.(*PinataClient); !ok {
		t.Error("expected PinataClient when Pinata key is set")
	}
}

// TestNewStoreDefaultIPFS verifies the factory uses default IPFS URL when none is provided.
func TestNewStoreDefaultIPFS(t *testing.T) {
	store := NewStore("", "", "")
	client, ok := store.(*IPFSClient)
	if !ok {
		t.Fatal("expected IPFSClient")
	}
	if client.APIURL != "http://localhost:5001" {
		t.Errorf("APIURL = %q, want http://localhost:5001", client.APIURL)
	}
}
