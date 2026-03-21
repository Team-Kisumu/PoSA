package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const maxUploadSize = 10 << 20 // 10MB

type SubmitResponse struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Size    int    `json:"size"`
	Message string `json:"message"`
}

type RepoRequest struct {
	Repo string `json:"repo"`
}

func Submit(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		handleFileUpload(w, r)
	case strings.HasPrefix(ct, "application/json"):
		handleRepoLink(w, r)
	default:
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{
			"error": "Content-Type must be multipart/form-data or application/json",
		})
	}
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{
			"error": "file exceeds 10MB limit",
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing or invalid 'file' field",
		})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to read file",
		})
		return
	}

	writeJSON(w, http.StatusOK, SubmitResponse{
		Type:    "file",
		Name:    header.Filename,
		Size:    len(content),
		Message: "file received, pending evaluation",
	})
}

func handleRepoLink(w http.ResponseWriter, r *http.Request) {
	var req RepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}

	if req.Repo == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "'repo' field is required",
		})
		return
	}

	if !strings.HasPrefix(req.Repo, "https://github.com/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "repo must be a valid GitHub URL (https://github.com/...)",
		})
		return
	}

	writeJSON(w, http.StatusOK, SubmitResponse{
		Type:    "repo",
		Name:    req.Repo,
		Size:    0,
		Message: "repo link received, pending evaluation",
	})
}
