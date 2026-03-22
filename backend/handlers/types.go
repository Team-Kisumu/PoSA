// Package handlers implements HTTP route handlers for the PoSA backend API.
// All responses use a consistent APIResponse envelope to simplify client-side
// parsing and error handling. Validation logic lives in the validation package.
package handlers

// --- Response Types ---

// APIResponse is the standard envelope for all API responses.
// On success: Success=true, Data is populated, Error is nil.
// On failure: Success=false, Data is nil, Error contains the code and message.
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error represents a structured error with a machine-readable code
// (e.g. "INVALID_CID") and a human-readable message.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HealthData is the response payload for the GET /health endpoint.
type HealthData struct {
	Status string `json:"status"`
}

// SubmitResponse is the response payload for successful POST /api/submit requests.
// Type indicates the submission kind ("file" or "repo").
// Size is the byte count of the uploaded file (0 for repo submissions).
// MIME is the detected content type of the uploaded file (empty for repo submissions).
type SubmitResponse struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	MIME    string `json:"mime,omitempty"`
	Message string `json:"message"`
}

// VerifyResponse is the response payload for GET /api/verify/{cid}.
type VerifyResponse struct {
	CID     string `json:"cid"`
	Message string `json:"message"`
}

// --- Request Types ---

// RepoRequest is the expected JSON body for repo-link submissions.
// Only the "repo" field is accepted; unknown fields are rejected
// via DisallowUnknownFields() in the decoder.
type RepoRequest struct {
	Repo string `json:"repo"`
}
