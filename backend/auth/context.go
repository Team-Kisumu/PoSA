package auth

import (
	"context"
	"net/http"

	"github.com/Murzuqisah/PoSA/database"
)

type contextKey string

const userContextKey contextKey = "user"

// setUser attaches the authenticated user to the request context.
func setUser(r *http.Request, user *database.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// GetUser retrieves the authenticated user from the request context.
// Returns nil if no user is authenticated.
func GetUser(r *http.Request) *database.User {
	user, _ := r.Context().Value(userContextKey).(*database.User)
	return user
}
