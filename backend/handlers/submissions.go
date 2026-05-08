package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Murzuqisah/PoSA/auth"
	"github.com/Murzuqisah/PoSA/database"
)

// SubmissionsHandler handles saving evaluation results to the database.
type SubmissionsHandler struct {
	DB *database.DB
}

// Create handles POST /api/submissions — saves an evaluation result for the authenticated user.
func (h *SubmissionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var body struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Score  int    `json:"score"`
		CID    string `json:"cid"`
		TxHash string `json:"tx_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if body.Name == "" {
		respondError(w, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}
	if body.Type == "" {
		body.Type = "file"
	}
	if body.Score < 0 || body.Score > 100 {
		respondError(w, http.StatusBadRequest, "INVALID_SCORE", "score must be 0-100")
		return
	}

	sub, err := h.DB.CreateSubmission(user.ID, body.Type, body.Name, body.Score, body.CID, body.TxHash)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to save submission")
		return
	}

	respondOK(w, sub)
}
