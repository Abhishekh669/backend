package router

import (
	"net/http"

	"github.com/Abhishekh669/backend/controller"
	"github.com/gorilla/mux"
)

func AuthRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/auth/fuck", func(w http.ResponseWriter, r *http.Request) {
		// Set the Content-Type to text/html
		w.Header().Set("Content-Type", "text/html")

		// Write the HTML response
		w.Write([]byte("<h1>Hello W hdjlorld f heckuck </h1>"))
	}).Methods("GET")
	router.HandleFunc("/api/auth/google", controller.AuthHandler).Methods("GET")
	router.HandleFunc("/api/auth/session", controller.GetValidateSession).Methods("GET")
	router.HandleFunc("/api/auth/google/callback", controller.AuthGoogleCallback).Methods("GET")

	return router
}
