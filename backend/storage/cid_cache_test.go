package storage

import (
	"fmt"
	"testing"
	"time"
)

// --- CID Generation Tests ---

// TestComputeCID verifies deterministic CID generation from content.
func TestComputeCID(t *testing.T) {
	data := []byte(`{"score":85,"issues":[]}`)

	cid1, err := ComputeCID(data)
	if err != nil {
		t.Fatalf("ComputeCID failed: %v", err)
	}
	if cid1 == "" {
		t.Fatal("empty CID returned")
	}

	// Same content should produce the same CID (deterministic).
	cid2, err := ComputeCID(data)
	if err != nil {
		t.Fatalf("ComputeCID failed: %v", err)
	}
	if cid1 != cid2 {
		t.Errorf("CIDs differ for same content: %q vs %q", cid1, cid2)
	}
	t.Logf("CID: %s", cid1)
}

// TestComputeCIDDifferentContent verifies different content produces different CIDs.
func TestComputeCIDDifferentContent(t *testing.T) {
	cid1, _ := ComputeCID([]byte(`{"score":85}`))
	cid2, _ := ComputeCID([]byte(`{"score":86}`))

	if cid1 == cid2 {
		t.Error("different content produced the same CID")
	}
}

// TestComputeCIDEmpty verifies empty data is rejected.
func TestComputeCIDEmpty(t *testing.T) {
	_, err := ComputeCID([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

// TestValidateCID verifies CID validation accepts valid CIDs and rejects invalid ones.
func TestValidateCID(t *testing.T) {
	// Generate a valid CID first.
	validCID, _ := ComputeCID([]byte("test content"))

	normalized, err := ValidateCID(validCID)
	if err != nil {
		t.Fatalf("ValidateCID(%q) failed: %v", validCID, err)
	}
	if normalized == "" {
		t.Error("empty normalized CID")
	}
}

// TestValidateCIDEmpty verifies empty CID is rejected.
func TestValidateCIDEmpty(t *testing.T) {
	_, err := ValidateCID("")
	if err == nil {
		t.Error("expected error for empty CID")
	}
}

// TestValidateCIDInvalid verifies malformed CIDs are rejected.
func TestValidateCIDInvalid(t *testing.T) {
	_, err := ValidateCID("not-a-valid-cid")
	if err == nil {
		t.Error("expected error for invalid CID")
	}
}

// TestVerifyContent verifies content integrity checking.
func TestVerifyContent(t *testing.T) {
	data := []byte(`{"score":92}`)
	cidStr, _ := ComputeCID(data)

	match, err := VerifyContent(cidStr, data)
	if err != nil {
		t.Fatalf("VerifyContent failed: %v", err)
	}
	if !match {
		t.Error("expected content to match its CID")
	}
}

// TestVerifyContentMismatch verifies tampered content is detected.
func TestVerifyContentMismatch(t *testing.T) {
	original := []byte(`{"score":92}`)
	cidStr, _ := ComputeCID(original)

	tampered := []byte(`{"score":100}`)
	match, err := VerifyContent(cidStr, tampered)
	if err != nil {
		t.Fatalf("VerifyContent failed: %v", err)
	}
	if match {
		t.Error("expected mismatch for tampered content")
	}
}

// --- Cache Tests ---

// mockStore is a simple in-memory Store for testing the cache layer.
type mockStore struct {
	uploads   map[string][]byte
	callCount int
}

func newMockStore() *mockStore {
	return &mockStore{uploads: make(map[string][]byte)}
}

func (m *mockStore) Upload(report []byte) (string, error) {
	m.callCount++
	cid, _ := ComputeCID(report)
	m.uploads[cid] = report
	return cid, nil
}

func (m *mockStore) Retrieve(cid string) ([]byte, error) {
	m.callCount++
	data, ok := m.uploads[cid]
	if !ok {
		return nil, fmt.Errorf("not found: %s", cid)
	}
	return data, nil
}

// TestCachedStoreUpload verifies upload caches the result.
func TestCachedStoreUpload(t *testing.T) {
	mock := newMockStore()
	cached := NewCachedStore(mock, nil)

	report := []byte(`{"score":85}`)
	cid, err := cached.Upload(report)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if cid == "" {
		t.Fatal("empty CID")
	}
	if cached.Size() != 1 {
		t.Errorf("cache size = %d, want 1", cached.Size())
	}
}

// TestCachedStoreRetrieveCacheHit verifies cache hit avoids store call.
func TestCachedStoreRetrieveCacheHit(t *testing.T) {
	mock := newMockStore()
	cached := NewCachedStore(mock, nil)

	report := []byte(`{"score":85}`)
	cid, _ := cached.Upload(report)

	// Reset call count after upload.
	mock.callCount = 0

	// Retrieve should hit cache, not the store.
	data, err := cached.Retrieve(cid)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if string(data) != string(report) {
		t.Errorf("data mismatch")
	}
	if mock.callCount != 0 {
		t.Errorf("store was called %d times, want 0 (cache hit)", mock.callCount)
	}
}

// TestCachedStoreRetrieveCacheMiss verifies cache miss calls the store.
func TestCachedStoreRetrieveCacheMiss(t *testing.T) {
	mock := newMockStore()
	cached := NewCachedStore(mock, nil)

	// Upload directly to mock (bypassing cache).
	report := []byte(`{"score":90}`)
	cid, _ := mock.Upload(report)
	mock.callCount = 0

	// Retrieve should miss cache and call the store.
	data, err := cached.Retrieve(cid)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if string(data) != string(report) {
		t.Errorf("data mismatch")
	}
	if mock.callCount != 1 {
		t.Errorf("store was called %d times, want 1 (cache miss)", mock.callCount)
	}

	// Second retrieve should hit cache.
	mock.callCount = 0
	_, _ = cached.Retrieve(cid)
	if mock.callCount != 0 {
		t.Errorf("store was called on second retrieve, want cache hit")
	}
}

// TestCachedStoreEviction verifies LRU eviction when cache is full.
func TestCachedStoreEviction(t *testing.T) {
	mock := newMockStore()
	cached := NewCachedStore(mock, &CacheConfig{MaxSize: 2})

	// Fill cache with 2 entries.
	cached.Upload([]byte(`{"id":1}`))
	cached.Upload([]byte(`{"id":2}`))
	if cached.Size() != 2 {
		t.Fatalf("cache size = %d, want 2", cached.Size())
	}

	// Third upload should evict the oldest.
	cached.Upload([]byte(`{"id":3}`))
	if cached.Size() != 2 {
		t.Errorf("cache size = %d, want 2 after eviction", cached.Size())
	}
}

// TestCachedStoreTTL verifies expired entries are not returned.
func TestCachedStoreTTL(t *testing.T) {
	mock := newMockStore()
	cached := NewCachedStore(mock, &CacheConfig{TTL: 50 * time.Millisecond})

	report := []byte(`{"score":85}`)
	cid, _ := cached.Upload(report)

	// Should hit cache immediately.
	mock.callCount = 0
	_, _ = cached.Retrieve(cid)
	if mock.callCount != 0 {
		t.Error("expected cache hit before TTL")
	}

	// Wait for TTL to expire.
	time.Sleep(60 * time.Millisecond)

	// Should miss cache after TTL.
	mock.callCount = 0
	_, _ = cached.Retrieve(cid)
	if mock.callCount != 1 {
		t.Errorf("expected cache miss after TTL, store called %d times", mock.callCount)
	}
}

// TestCachedStoreStats verifies stats output.
func TestCachedStoreStats(t *testing.T) {
	mock := newMockStore()
	cached := NewCachedStore(mock, &CacheConfig{MaxSize: 100, TTL: 5 * time.Minute})

	cached.Upload([]byte(`{"score":1}`))
	cached.Upload([]byte(`{"score":2}`))

	stats := cached.Stats()
	if stats == "" {
		t.Error("empty stats")
	}
	t.Logf("Stats: %s", stats)
}
