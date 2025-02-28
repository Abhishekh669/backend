package router

import (
	"github.com/Abhishekh669/backend/controller"
	"github.com/gorilla/mux"
)

func StudentAbsentRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/student/absent", controller.CreateStudentAbsentHandler).Methods("POST")
	router.HandleFunc("/api/student/absent/edit", controller.UpdateAbsentRequestHandler).Methods("POST")
	router.HandleFunc("/api/student/absent/all", controller.GetUserAllAbsentResultsHandler).Methods("GET")
	return router
}
