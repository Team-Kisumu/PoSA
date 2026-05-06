package database

import (
	"os"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	path := t.TempDir() + "/test.db"
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open(%s) failed: %v", path, err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// --- User Tests ---

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)

	user, err := db.CreateUser(12345, "testuser", "https://avatar.url", "test@example.com")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("username = %q, want %q", user.Username, "testuser")
	}
	if user.GitHubID != 12345 {
		t.Errorf("github_id = %d, want %d", user.GitHubID, 12345)
	}
	if user.Role != "user" {
		t.Errorf("role = %q, want %q", user.Role, "user")
	}
}

func TestCreateUserUpsert(t *testing.T) {
	db := setupTestDB(t)

	db.CreateUser(12345, "oldname", "", "")
	user, err := db.CreateUser(12345, "newname", "https://new.avatar", "new@email.com")
	if err != nil {
		t.Fatalf("CreateUser upsert failed: %v", err)
	}
	if user.Username != "newname" {
		t.Errorf("username = %q, want %q after upsert", user.Username, "newname")
	}
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)

	created, _ := db.CreateUser(99, "byid", "", "")
	found, err := db.GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if found == nil || found.Username != "byid" {
		t.Error("GetUserByID returned wrong user")
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	db := setupTestDB(t)

	user, err := db.GetUserByID(999)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user != nil {
		t.Error("expected nil for non-existent user")
	}
}

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)

	db.CreateUser(1, "user1", "", "")
	db.CreateUser(2, "user2", "", "")
	db.CreateUser(3, "user3", "", "")

	users, err := db.ListUsers(10, 0)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("len(users) = %d, want 3", len(users))
	}
}

func TestListUsersPagination(t *testing.T) {
	db := setupTestDB(t)

	for i := 1; i <= 5; i++ {
		db.CreateUser(int64(i), "user", "", "")
	}

	users, _ := db.ListUsers(2, 0)
	if len(users) != 2 {
		t.Errorf("page 1: len = %d, want 2", len(users))
	}

	users, _ = db.ListUsers(2, 4)
	if len(users) != 1 {
		t.Errorf("page 3: len = %d, want 1", len(users))
	}
}

func TestUpdateUserRole(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.CreateUser(1, "admin", "", "")
	db.UpdateUserRole(user.ID, "admin")

	updated, _ := db.GetUserByID(user.ID)
	if updated.Role != "admin" {
		t.Errorf("role = %q, want %q", updated.Role, "admin")
	}
}

func TestCountUsers(t *testing.T) {
	db := setupTestDB(t)

	db.CreateUser(1, "a", "", "")
	db.CreateUser(2, "b", "", "")

	count, err := db.CountUsers()
	if err != nil {
		t.Fatalf("CountUsers failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// --- Session Tests ---

func TestCreateAndGetSession(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.CreateUser(1, "sessionuser", "", "")
	token, err := db.CreateSession(user.ID, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("token length = %d, want 64", len(token))
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

	user, _ := db.GetSession("nonexistent_token")
	if user != nil {
		t.Error("expected nil for invalid token")
	}
}

func TestGetSessionExpired(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.CreateUser(1, "expired", "", "")
	token, _ := db.CreateSession(user.ID, -1*time.Hour) // Already expired.

	found, _ := db.GetSession(token)
	if found != nil {
		t.Error("expected nil for expired session")
	}
}

func TestDeleteSession(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.CreateUser(1, "logout", "", "")
	token, _ := db.CreateSession(user.ID, 24*time.Hour)

	db.DeleteSession(token)

	found, _ := db.GetSession(token)
	if found != nil {
		t.Error("session should be deleted")
	}
}

func TestDeleteUserSessions(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.CreateUser(1, "multi", "", "")
	db.CreateSession(user.ID, 24*time.Hour)
	db.CreateSession(user.ID, 24*time.Hour)

	db.DeleteUserSessions(user.ID)

	// Both sessions should be gone — create a new one to verify the user still works.
	token, _ := db.CreateSession(user.ID, 24*time.Hour)
	found, _ := db.GetSession(token)
	if found == nil {
		t.Error("new session after delete should work")
	}
}

// --- Submission Tests ---

func TestCreateSubmission(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.CreateUser(1, "submitter", "", "")
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

	user, _ := db.CreateUser(1, "lister", "", "")
	db.CreateSubmission(user.ID, "file", "a.go", 90, "", "")
	db.CreateSubmission(user.ID, "repo", "https://github.com/x/y", 75, "", "")

	subs, err := db.ListSubmissions(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListSubmissions failed: %v", err)
	}
	if len(subs) != 2 {
		t.Errorf("len = %d, want 2", len(subs))
	}
}

func TestListSubmissionsAll(t *testing.T) {
	db := setupTestDB(t)

	u1, _ := db.CreateUser(1, "u1", "", "")
	u2, _ := db.CreateUser(2, "u2", "", "")
	db.CreateSubmission(u1.ID, "file", "a.go", 90, "", "")
	db.CreateSubmission(u2.ID, "file", "b.go", 80, "", "")

	subs, _ := db.ListSubmissions(0, 10, 0) // userID=0 means all.
	if len(subs) != 2 {
		t.Errorf("len = %d, want 2", len(subs))
	}
}

func TestCountSubmissions(t *testing.T) {
	db := setupTestDB(t)

	user, _ := db.CreateUser(1, "counter", "", "")
	db.CreateSubmission(user.ID, "file", "a.go", 90, "", "")
	db.CreateSubmission(user.ID, "file", "b.go", 80, "", "")

	count, _ := db.CountSubmissions(user.ID)
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	total, _ := db.CountSubmissions(0)
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
}

// --- Database Open Tests ---

func TestOpenCreatesDirectory(t *testing.T) {
	path := t.TempDir() + "/nested/dir/test.db"
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	db.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestOpenFromEnvVar(t *testing.T) {
	path := t.TempDir() + "/env.db"
	t.Setenv("DATABASE_URL", path)

	db, err := Open("")
	if err != nil {
		t.Fatalf("Open from env failed: %v", err)
	}
	db.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("database file was not created from DATABASE_URL")
	}
}
