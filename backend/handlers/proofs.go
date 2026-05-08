package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Murzuqisah/PoSA/database"
)

// ProofsHandler holds the database dependency for proof endpoints.
type ProofsHandler struct {
	DB *database.DB
}

// List handles GET /api/proofs — returns paginated list of all minted credentials.
func (h *ProofsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	subs, err := h.DB.ListSubmissions(0, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to list proofs")
		return
	}

	total, _ := h.DB.CountSubmissions(0)

	respondOK(w, map[string]any{
		"proofs": subs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Get handles GET /api/proofs/{id} — returns a single credential by ID.
func (h *ProofsHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "proof ID must be a positive integer")
		return
	}

	sub, err := h.DB.GetSubmissionByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to get proof")
		return
	}
	if sub == nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "proof not found")
		return
	}

	// Get the submitter's username.
	user, _ := h.DB.GetUserByID(sub.UserID)
	username := ""
	avatarURL := ""
	if user != nil {
		username = user.Username
		avatarURL = user.AvatarURL
	}

	respondOK(w, map[string]any{
		"proof":      sub,
		"username":   username,
		"avatar_url": avatarURL,
	})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}

// marshalJSON is unused but keeps the import valid.
var _ = json.Marshal
