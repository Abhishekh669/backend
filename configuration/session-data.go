package configuration

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// SessionType represents user session data.
type SessionType struct {
	UserId        string `json:"userId,omitempty"`
	Email         string `json:"email,omitempty"`
	Name          string `json:"name,omitempty"`
	Authenticated bool   `json:"authenticated"`
}

// AdminSessionType represents admin session data.
type AdminSessionType struct {
	AdminId       string `json:"adminId,omitempty"`
	UserId        string `json:"userId,omitempty"`
	Email         string `json:"email,omitempty"`
	Authenticated bool   `json:"authenticated"`
}

// GetSessionData retrieves user session data from the request.
func GetSessionData(r *http.Request) (SessionType, error) {
	sessionKey := []byte(os.Getenv("SESSION_SECRET"))
	if len(sessionKey) == 0 {
		return SessionType{}, fmt.Errorf("session secret is not set")
	}

	sessionStore := sessions.NewCookieStore(sessionKey)
	session, err := sessionStore.Get(r, "__session")
	if err != nil {
		return SessionType{}, fmt.Errorf("failed to get session: %v", err)
	}

	// Check if required session values exist
	userId, ok := session.Values["user_id"].(string)
	if !ok {
		return SessionType{}, fmt.Errorf("missing or invalid user_id in session")
	}
	email, ok := session.Values["email"].(string)
	if !ok {
		return SessionType{}, fmt.Errorf("missing or invalid email in session")
	}
	name, ok := session.Values["name"].(string)
	if !ok {
		name = "" // Optional field
	}
	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok {
		return SessionType{}, fmt.Errorf("missing or invalid authenticated flag in session")
	}

	return SessionType{
		UserId:        userId,
		Email:         email,
		Name:          name,
		Authenticated: authenticated,
	}, nil
}

// GetAdminSessionData retrieves admin session data from the request.
func GetAdminSessionData(r *http.Request) (AdminSessionType, error) {
	adminKey := []byte(os.Getenv("ADMIN_SECRET"))
	if len(adminKey) == 0 {
		return AdminSessionType{}, fmt.Errorf("admin secret is not set")
	}

	adminSessionStore := sessions.NewCookieStore(adminKey)
	session, err := adminSessionStore.Get(r, "admin_token")
	if err != nil {
		return AdminSessionType{}, fmt.Errorf("failed to get admin session: %v", err)
	}

	// Check if required session values exist
	adminId, ok := session.Values["adminId"].(string)
	if !ok {
		return AdminSessionType{}, fmt.Errorf("missing or invalid adminId in session")
	}
	userId, ok := session.Values["userId"].(string)
	if !ok {
		return AdminSessionType{}, fmt.Errorf("missing or invalid userId in session")
	}
	email, ok := session.Values["email"].(string)
	if !ok {
		return AdminSessionType{}, fmt.Errorf("missing or invalid email in session")
	}
	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok {
		return AdminSessionType{}, fmt.Errorf("missing or invalid authenticated flag in session")
	}

	return AdminSessionType{
		AdminId:       adminId,
		UserId:        userId,
		Email:         email,
		Authenticated: authenticated,
	}, nil
}
