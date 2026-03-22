package handlers

import (
	"encoding/json"
	"testing"
)

// TestAPIResponseEnvelopeJSON verifies that APIResponse serializes correctly
// for both success and error cases, producing the expected JSON structure.
func TestAPIResponseEnvelopeJSON(t *testing.T) {
	t.Run("success response omits error", func(t *testing.T) {
		resp := APIResponse{Success: true, Data: HealthData{Status: "ok"}}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		// Verify "error" key is omitted (omitempty) and "data" is present.
		var raw map[string]json.RawMessage
		json.Unmarshal(data, &raw)
		if _, ok := raw["error"]; ok {
			t.Error("success response should not contain 'error' key")
		}
		if _, ok := raw["data"]; !ok {
			t.Error("success response should contain 'data' key")
		}
	})

	t.Run("error response omits data", func(t *testing.T) {
		resp := APIResponse{Success: false, Error: &Error{Code: "TEST", Message: "test error"}}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		// Verify "data" key is omitted (omitempty) and "error" is present.
		var raw map[string]json.RawMessage
		json.Unmarshal(data, &raw)
		if _, ok := raw["data"]; ok {
			t.Error("error response should not contain 'data' key")
		}
		if _, ok := raw["error"]; !ok {
			t.Error("error response should contain 'error' key")
		}
	})
}

// TestSubmitResponseMIMEField verifies the MIME field is included when set
// and omitted when empty (omitempty behavior).
func TestSubmitResponseMIMEField(t *testing.T) {
	t.Run("with MIME", func(t *testing.T) {
		sr := SubmitResponse{Type: "file", Name: "main.go", Size: 100, MIME: "text/plain", Message: "ok"}
		data, _ := json.Marshal(sr)
		var raw map[string]json.RawMessage
		json.Unmarshal(data, &raw)
		if _, ok := raw["mime"]; !ok {
			t.Error("expected 'mime' field when MIME is set")
		}
	})

	t.Run("without MIME", func(t *testing.T) {
		sr := SubmitResponse{Type: "repo", Name: "https://github.com/user/repo", Size: 0, Message: "ok"}
		data, _ := json.Marshal(sr)
		var raw map[string]json.RawMessage
		json.Unmarshal(data, &raw)
		if _, ok := raw["mime"]; ok {
			t.Error("expected 'mime' field to be omitted when empty")
		}
	})
}
