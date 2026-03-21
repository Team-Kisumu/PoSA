package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func Submit(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		respondError(w, http.StatusBadRequest, "MISSING_CONTENT_TYPE", "Content-Type header is required")
		return
	}

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		handleFileUpload(w, r)
	case strings.HasPrefix(ct, "application/json"):
		handleRepoLink(w, r)
	default:
		respondError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE",
			"Content-Type must be multipart/form-data or application/json")
	}
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file exceeds 10MB limit")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "MISSING_FILE", "missing or invalid 'file' field")
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	if err := ValidateFilename(filename); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_FILENAME", err.Error())
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "READ_ERROR", "failed to read file")
		return
	}

	if len(content) == 0 {
		respondError(w, http.StatusBadRequest, "EMPTY_FILE", "file is empty")
		return
	}

	respondOK(w, SubmitResponse{
		Type:    "file",
		Name:    filename,
		Size:    int64(len(content)),
		Message: "file received, pending evaluation",
	})
}

func handleRepoLink(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req RepoRequest
	if err := decoder.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON body")
		return
	}

	req.Repo = strings.TrimSpace(req.Repo)
	if err := ValidateRepoURL(req.Repo); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REPO", err.Error())
		return
	}

	respondOK(w, SubmitResponse{
		Type:    "repo",
		Name:    req.Repo,
		Size:    0,
		Message: "repo link received, pending evaluation",
	})
}
