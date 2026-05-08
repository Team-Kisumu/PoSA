package database

import (
	"os"
	"strings"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	dsn := os.Getenv("SUPABASE_DB_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("SUPABASE_DB_URL or DATABASE_URL not set — skipping database tests")
	}
	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// cleanupTestUser removes a test user by github_id.
func cleanupTestUser(t *testing.T, db *DB, githubID int64) {
	t.Helper()
	user, _ := db.GetUserByGitHubID(githubID)
	if user != nil {
		db.conn.Exec(`DELETE FROM submissions WHERE user_id = $1`, user.ID)
		db.conn.Exec(`DELETE FROM sessions WHERE user_id = $1`, user.ID)
		db.conn.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	}
}

// --- User Tests ---

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900001)

	user, err := db.CreateUser(900001, "testuser_create", "https://avatar.url", "test@example.com")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if user.Username != "testuser_create" {
		t.Errorf("username = %q, want %q", user.Username, "testuser_create")
	}
	if user.GitHubID != 900001 {
		t.Errorf("github_id = %d, want %d", user.GitHubID, 900001)
	}
	if user.Role != "user" {
		t.Errorf("role = %q, want %q", user.Role, "user")
	}
}

func TestCreateUserUpsert(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900002)

	db.CreateUser(900002, "oldname", "", "")
	user, err := db.CreateUser(900002, "newname", "https://new.avatar", "new@email.com")
	if err != nil {
		t.Fatalf("CreateUser upsert failed: %v", err)
	}
	if user.Username != "newname" {
		t.Errorf("username = %q, want %q after upsert", user.Username, "newname")
	}
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900003)

	created, _ := db.CreateUser(900003, "byid_test", "", "")
	found, err := db.GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if found == nil || found.Username != "byid_test" {
		t.Error("GetUserByID returned wrong user")
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	db := setupTestDB(t)

	user, err := db.GetUserByID(999999)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user != nil {
		t.Error("expected nil for non-existent user")
	}
}

func TestUpdateUserRole(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900004)

	user, _ := db.CreateUser(900004, "role_test", "", "")
	db.UpdateUserRole(user.ID, "admin")

	updated, _ := db.GetUserByID(user.ID)
	if updated.Role != "admin" {
		t.Errorf("role = %q, want %q", updated.Role, "admin")
	}
}

// --- Session Tests ---

func TestCreateAndGetSession(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900010)

	user, _ := db.CreateUser(900010, "session_test", "", "")
	token, err := db.CreateSession(user.ID, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if len(token) < 64 {
		t.Errorf("token length = %d, want >= 64", len(token))
	}

	found, err := db.GetSession(token)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if found == nil || found.ID != user.ID {
		t.Error("GetSession returned wrong user")
	}
}

func TestGetSessionInvalidToken(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.GetSession("nonexistent_token_that_is_long_enough")
	if user != nil {
		t.Error("expected nil for invalid token")
	}
}

func TestGetSessionExpired(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900011)

	user, _ := db.CreateUser(900011, "expired_test", "", "")
	token, _ := db.CreateSession(user.ID, -1*time.Hour)

	found, _ := db.GetSession(token)
	if found != nil {
		t.Error("expected nil for expired session")
	}
}

func TestDeleteSession(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900012)

	user, _ := db.CreateUser(900012, "delete_session_test", "", "")
	token, _ := db.CreateSession(user.ID, 24*time.Hour)

	db.DeleteSession(token)

	found, _ := db.GetSession(token)
	if found != nil {
		t.Error("session should be deleted")
	}
}

// --- Submission Tests ---

func TestCreateSubmission(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900020)

	user, _ := db.CreateUser(900020, "submit_test", "", "")
	sub, err := db.CreateSubmission(user.ID, "file", "main.go", 85, "QmTest123", "0xabc")
	if err != nil {
		t.Fatalf("CreateSubmission failed: %v", err)
	}
	if sub.Score != 85 {
		t.Errorf("score = %d, want 85", sub.Score)
	}
	if sub.CID != "QmTest123" {
		t.Errorf("cid = %q, want %q", sub.CID, "QmTest123")
	}
}

func TestListSubmissions(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900021)

	user, _ := db.CreateUser(900021, "list_sub_test", "", "")
	db.CreateSubmission(user.ID, "file", "a.go", 90, "", "")
	db.CreateSubmission(user.ID, "repo", "https://github.com/x/y", 75, "", "")

	subs, err := db.ListSubmissions(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListSubmissions failed: %v", err)
	}
	if len(subs) < 2 {
		t.Errorf("len = %d, want >= 2", len(subs))
	}
}

// --- HMAC Token Tests ---

func TestValidateTokenNoSecret(t *testing.T) {
	sessionSecret = ""
	if !ValidateToken("anyrandomtoken") {
		t.Error("expected valid in dev mode")
	}
}

func TestValidateTokenWithSecret(t *testing.T) {
	sessionSecret = "testsecret123"
	defer func() { sessionSecret = "" }()

	token := generateSessionToken()
	if !ValidateToken(token) {
		t.Error("valid signed token rejected")
	}

	if ValidateToken("tampered" + token[8:]) {
		t.Error("tampered token accepted")
	}

	if ValidateToken("plainunsignedtoken") {
		t.Error("unsigned token accepted when secret is set")
	}
}

func TestSessionWithHMAC(t *testing.T) {
	sessionSecret = "hmactest"
	defer func() { sessionSecret = "" }()

	db := setupTestDB(t)
	defer cleanupTestUser(t, db, 900030)

	user, _ := db.CreateUser(900030, "hmac_test", "", "")
	token, err := db.CreateSession(user.ID, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if !strings.Contains(token, ".") {
		t.Error("signed token should contain a dot separator")
	}

	found, err := db.GetSession(token)
	if err != nil || found == nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if found.Username != "hmac_test" {
		t.Errorf("username = %q, want %q", found.Username, "hmac_test")
	}

	found, _ = db.GetSession("tampered." + token[strings.IndexByte(token, '.')+1:])
	if found != nil {
		t.Error("tampered token should not return a user")
	}
}
