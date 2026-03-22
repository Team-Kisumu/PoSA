package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Murzuqisah/PoSA/validation"
)

// Health handles GET /health and returns a simple status check.
// Used by load balancers and monitoring to confirm the server is running.
func Health(w http.ResponseWriter, r *http.Request) {
	respondOK(w, HealthData{Status: "ok"})
}

// Verify handles GET /api/verify/{cid}. It extracts the CID path parameter,
// validates it against injection and format rules, and returns the verification
// result. Currently returns a placeholder until blockchain lookup is implemented.
func Verify(w http.ResponseWriter, r *http.Request) {
	cid := r.PathValue("cid")
	if err := validation.CID(cid); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_CID", err.Error())
		return
	}
	respondOK(w, VerifyResponse{
		CID:     cid,
		Message: "verify endpoint not yet implemented",
	})
}

// respondOK sends a 200 JSON response with the given data wrapped
// in the standard APIResponse envelope (success=true).
func respondOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

// respondError sends an error JSON response with the given HTTP status,
// machine-readable error code, and human-readable message.
func respondError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, APIResponse{Success: false, Error: &Error{Code: code, Message: message}})
}

// writeJSON serializes the APIResponse as JSON and writes it to the
// response writer with the appropriate Content-Type header and status code.
func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
