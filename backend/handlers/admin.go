package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Murzuqisah/PoSA/database"
)

// AdminHandler holds the database dependency for admin endpoints.
type AdminHandler struct {
	DB *database.DB
}

// ListUsers handles GET /api/admin/users — paginated user list.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	if limit > 100 {
		limit = 100
	}

	users, err := h.DB.ListUsers(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to list users")
		return
	}
	total, _ := h.DB.CountUsers()

	respondOK(w, map[string]any{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUser handles GET /api/admin/users/{id} — user detail with their submissions.
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "user ID must be a positive integer")
		return
	}

	user, err := h.DB.GetUserByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to get user")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}

	subs, _ := h.DB.ListSubmissions(id, 20, 0)
	subCount, _ := h.DB.CountSubmissions(id)

	respondOK(w, map[string]any{
		"user":              user,
		"submissions":       subs,
		"total_submissions": subCount,
	})
}

// UpdateUser handles PATCH /api/admin/users/{id} — update user role.
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "user ID must be a positive integer")
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if body.Role != "user" && body.Role != "admin" {
		respondError(w, http.StatusBadRequest, "INVALID_ROLE", "role must be 'user' or 'admin'")
		return
	}

	if err := h.DB.UpdateUserRole(id, body.Role); err != nil {
		respondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to update user")
		return
	}

	user, _ := h.DB.GetUserByID(id)
	respondOK(w, user)
}

// ListCredentials handles GET /api/admin/credentials — all submissions.
func (h *AdminHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	if limit > 100 {
		limit = 100
	}

	subs, err := h.DB.ListSubmissions(0, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to list credentials")
		return
	}
	total, _ := h.DB.CountSubmissions(0)

	respondOK(w, map[string]any{
		"credentials": subs,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// Stats handles GET /api/admin/stats — aggregate statistics.
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	totalUsers, _ := h.DB.CountUsers()
	totalSubmissions, _ := h.DB.CountSubmissions(0)

	respondOK(w, map[string]any{
		"total_users":       totalUsers,
		"total_submissions": totalSubmissions,
	})
}
