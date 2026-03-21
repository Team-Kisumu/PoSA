package handlers

import (
	"encoding/json"
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func Verify(w http.ResponseWriter, r *http.Request) {
	cid := r.PathValue("cid")
	if cid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cid is required"})
		return
	}
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "verify endpoint not yet implemented",
		"cid":     cid,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
