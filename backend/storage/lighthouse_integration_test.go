// go:build integration

package storage

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func loadLighthouseKey(t *testing.T) string {
	key := os.Getenv("LIGHTHOUSE_API_KEY")
	if key != "" {
		return key
	}
	// Fallback: read from .env in project root.
	data, err := os.ReadFile("../../.env")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "LIGHTHOUSE_API_KEY=") {
				key = strings.TrimPrefix(line, "LIGHTHOUSE_API_KEY=")
				if key != "" {
					return key
				}
			}
		}
	}
	t.Skip("LIGHTHOUSE_API_KEY not set — skipping Lighthouse integration tests")
	return ""
}

// TestLighthouseUploadLive uploads a real report to Lighthouse and verifies the CID.
func TestLighthouseUploadLive(t *testing.T) {
	apiKey := loadLighthouseKey(t)

	client := NewLighthouseClient(LighthouseConfig{APIKey: apiKey})

	report, _ := json.Marshal(map[string]any{
		"score":  92,
		"issues": []any{},
		"test":   "lighthouse-integration",
	})

	cid, err := client.Upload(report)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if cid == "" {
		t.Fatal("empty CID returned")
	}
	if !strings.HasPrefix(cid, "Qm") && !strings.HasPrefix(cid, "bafy") {
		t.Errorf("unexpected CID format: %s", cid)
	}
	t.Logf("Uploaded: CID=%s (%d bytes)", cid, len(report))
}

// TestLighthouseRoundtrip uploads a report and retrieves it back, verifying content matches.
func TestLighthouseRoundtrip(t *testing.T) {
	apiKey := loadLighthouseKey(t)

	client := NewLighthouseClient(LighthouseConfig{APIKey: apiKey})

	original := map[string]any{
		"score":    78,
		"filename": "roundtrip_test.go",
		"issues":   []string{"test issue"},
	}
	reportJSON, _ := json.Marshal(original)

	// Upload
	cid, err := client.Upload(reportJSON)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	t.Logf("Uploaded: CID=%s", cid)

	// Retrieve
	data, err := client.Retrieve(cid)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	// Verify content matches
	var retrieved map[string]any
	if err := json.Unmarshal(data, &retrieved); err != nil {
		t.Fatalf("Failed to parse retrieved data: %v", err)
	}
	if int(retrieved["score"].(float64)) != 78 {
		t.Errorf("score = %v, want 78", retrieved["score"])
	}
	if retrieved["filename"] != "roundtrip_test.go" {
		t.Errorf("filename = %v, want roundtrip_test.go", retrieved["filename"])
	}
	t.Logf("Roundtrip verified: %d bytes uploaded, %d bytes retrieved", len(reportJSON), len(data))
	t.Logf("Public URL: https://gateway.lighthouse.storage/ipfs/%s", cid)
}
