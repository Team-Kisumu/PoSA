package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Murzuqisah/PoSA/database"
)

func TestGetUserNil(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if user := GetUser(req); user != nil {
		t.Error("expected nil user on fresh request")
	}
}

func TestSetAndGetUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &database.User{ID: 1, Username: "test"}
	req = setUser(req, user)

	got := GetUser(req)
	if got == nil || got.Username != "test" {
		t.Errorf("GetUser = %v, want user with username 'test'", got)
	}
}

func TestRequireAuthRejects(t *testing.T) {
	handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRequireAuthAllows(t *testing.T) {
	handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = setUser(req, &database.User{ID: 1, Username: "authed"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRequireAdminRejectsUser(t *testing.T) {
	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = setUser(req, &database.User{ID: 1, Role: "user"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestRequireAdminAllows(t *testing.T) {
	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = setUser(req, &database.User{ID: 1, Role: "admin"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestMeUnauthenticated(t *testing.T) {
	db := setupTestDB(t)
	h := &Handler{DB: db}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	h.Me(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestLogout(t *testing.T) {
	db := setupTestDB(t)
	h := &Handler{DB: db}

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	h.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("SUPABASE_DB_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("SUPABASE_DB_URL not set — skipping DB-dependent auth test")
	}
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
