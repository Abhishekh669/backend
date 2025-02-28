package middleware

import (
	"net/http"

	"github.com/Abhishekh669/backend/configuration"
)

func ValidateSession(w http.ResponseWriter, r *http.Request) bool {
	sessionData, err := configuration.GetSessionData(r)
	if err != nil {
		return false

	}
	if !sessionData.Authenticated {
		return false
	}
	return true
}

func ValidateAdmin(w http.ResponseWriter, r *http.Request) bool {
	adminSessionData, err := configuration.GetAdminSessionData(r)
	if err != nil {
		return false
	}
	if !adminSessionData.Authenticated {
		return false
	}
	return true

}
