package router

import (
	"net/http"

	"github.com/Abhishekh669/backend/controller"
	"github.com/gorilla/mux"
)

func UserRouter() *mux.Router {

	router := mux.NewRouter()
	router.HandleFunc("/api/user", controller.CreateUserHandler).Methods("POST")
	router.HandleFunc("/api/user/all", controller.GetAllUsersHandler).Methods("GET")
	router.HandleFunc("/api/user/onboard", controller.OnboardingUserHandler).Methods("POST")
	router.HandleFunc("/api/user/{id}", controller.GetUserByIdHandler).Methods("GET")
	router.HandleFunc("/api/user/fuck", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to HTML
		w.Header().Set("Content-Type", "text/html")

		// Write an HTML response
		htmlResponse := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Hello Page</title>
			</head>
			<body>
				<h1>Hello, World!</h1>
				<p>mother fucker</p>
			</body>
			</html>
		`

		// Write the HTML response to the client
		w.Write([]byte(htmlResponse))
	}).Methods("GET")

	return router
}
