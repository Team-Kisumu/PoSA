package handlers

import (
	"encoding/json"
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	respondOK(w, HealthData{Status: "ok"})
}

func Verify(w http.ResponseWriter, r *http.Request) {
	cid := r.PathValue("cid")
	if err := ValidateCID(cid); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_CID", err.Error())
		return
	}
	respondOK(w, VerifyResponse{
		CID:     cid,
		Message: "verify endpoint not yet implemented",
	})
}

func respondOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, APIResponse{Success: false, Error: &Error{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
