// Package auth implements GitHub OAuth authentication for PoSA.
// Flow: /auth/github -> GitHub authorize -> /auth/github/callback -> session cookie -> /auth/me
package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Murzuqisah/PoSA/database"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"
	sessionCookieName  = "posa_session"
	sessionDuration    = 7 * 24 * time.Hour // 7 days
)

// Handler holds dependencies for auth endpoints.
type Handler struct {
	DB           *database.DB
	ClientID     string
	ClientSecret string
}

// NewHandler creates an auth handler from environment variables.
func NewHandler(db *database.DB) *Handler {
	return &Handler{
		DB:           db,
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	}
}

// GitHubLogin redirects the user to GitHub's OAuth authorize page.
func (h *Handler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	if h.ClientID == "" {
		http.Error(w, `{"success":false,"error":{"code":"AUTH_NOT_CONFIGURED","message":"GitHub OAuth not configured"}}`, http.StatusServiceUnavailable)
		return
	}
	redirectURI := getRedirectURI(r)
	url := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=read:user user:email",
		githubAuthorizeURL, h.ClientID, redirectURI)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GitHubCallback handles the OAuth callback from GitHub.
// Exchanges the code for an access token, fetches user info, creates/updates
// the user in the database, creates a session, and sets the session cookie.
func (h *Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"success":false,"error":{"code":"MISSING_CODE","message":"authorization code missing"}}`, http.StatusBadRequest)
		return
	}

	// Exchange code for access token.
	token, err := h.exchangeCode(code, getRedirectURI(r))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false,"error":{"code":"TOKEN_EXCHANGE_FAILED","message":"%s"}}`, err.Error()), http.StatusBadGateway)
		return
	}

	// Fetch GitHub user profile.
	ghUser, err := fetchGitHubUser(token)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false,"error":{"code":"USER_FETCH_FAILED","message":"%s"}}`, err.Error()), http.StatusBadGateway)
		return
	}

	// Create or update user in database.
	user, err := h.DB.CreateUser(ghUser.ID, ghUser.Login, ghUser.AvatarURL, ghUser.Email)
	if err != nil {
		http.Error(w, `{"success":false,"error":{"code":"DB_ERROR","message":"failed to save user"}}`, http.StatusInternalServerError)
		return
	}

	// Create session.
	sessionToken, err := h.DB.CreateSession(user.ID, sessionDuration)
	if err != nil {
		http.Error(w, `{"success":false,"error":{"code":"SESSION_ERROR","message":"failed to create session"}}`, http.StatusInternalServerError)
		return
	}

	// Set session cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	// Redirect to analyze page after successful login.
	http.Redirect(w, r, "/analyze", http.StatusTemporaryRedirect)
}

// Me returns the current authenticated user's profile.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success":false,"error":{"code":"UNAUTHORIZED","message":"not authenticated"}}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true, "data": user})
}

// Logout destroys the current session and clears the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		h.DB.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true,"data":{"message":"logged out"}}`))
}

// AuthMiddleware validates the session cookie and attaches the user to the request context.
// If no valid session exists, the request continues without a user (for optional auth).
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil && cookie.Value != "" {
			user, _ := h.DB.GetSession(cookie.Value)
			if user != nil {
				r = setUser(r, user)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth is middleware that rejects unauthenticated requests with 401.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUser(r) == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"success":false,"error":{"code":"UNAUTHORIZED","message":"authentication required"}}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin is middleware that rejects non-admin users with 403.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user == nil || user.Role != "admin" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"success":false,"error":{"code":"FORBIDDEN","message":"admin access required"}}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- GitHub API helpers ---

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

func (h *Handler) exchangeCode(code, redirectURI string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(http.MethodPost, githubTokenURL, nil)
	q := req.URL.Query()
	q.Set("client_id", h.ClientID)
	q.Set("client_secret", h.ClientSecret)
	q.Set("code", code)
	q.Set("redirect_uri", redirectURI)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		return "", fmt.Errorf("github: %s", result.Error)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}
	return result.AccessToken, nil
}

func fetchGitHubUser(token string) (*githubUser, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, githubUserURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API %d: %s", resp.StatusCode, string(body))
	}

	var user githubUser
	json.NewDecoder(resp.Body).Decode(&user)
	if user.ID == 0 {
		return nil, fmt.Errorf("invalid user response")
	}
	return &user, nil
}

func getRedirectURI(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/auth/github/callback", scheme, r.Host)
}
