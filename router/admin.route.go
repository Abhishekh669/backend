package router

import (
	"github.com/Abhishekh669/backend/controller"
	"github.com/gorilla/mux"
)

func AdminRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/admin/new", controller.AdminCreateHandler).Methods("POST")
	router.HandleFunc("/api/admin/all", controller.GetAllAdminHandler).Methods("GET")
	router.HandleFunc("/api/admin/login", controller.LoginHandler).Methods("POST")
	router.HandleFunc("/api/admin/token", controller.GetAdminFromTokenHandler).Methods("GET")
	router.HandleFunc("/api/admin/unverified/all", controller.GetUnVerifiedUserHandler).Methods("GET")
	router.HandleFunc("/api/admin/verify/{id}", controller.VerifyUserHandler).Methods("GET")
	router.HandleFunc("/api/admin/reject/{id}", controller.RejectUserHandler).Methods("GET")
	router.HandleFunc("/api/admin/users/setoremove", controller.SetOrRemoveAdminHandler).Methods("POST")
	router.HandleFunc("/api/admin/users/delete/{id}", controller.DeleteUserHandler).Methods("DELETE")
	router.HandleFunc("/api/admin/attendance", controller.CreateAttendanceTrackerHandler).Methods("POST")
	router.HandleFunc("/api/admin/attendance/update", controller.UpdateAttendanceTrackerHandler).Methods("POST")
	router.HandleFunc("/api/admin/attendance/today", controller.GetTodayAttendanceHandler).Methods("GET")
	router.HandleFunc("/api/admin/absent/all", controller.GetAllAbsentResultsHandler).Methods("GET")
	router.HandleFunc("/api/admin/absent/user/{id}", controller.GetUserAbsentByIdHandler).Methods("GET")
	router.HandleFunc("/api/admin/absent/user/reject", controller.RejectAbsentRequestHandler).Methods("POST")
	router.HandleFunc("/api/admin/absent/user/accept", controller.AcceptAbsentRequestHandler).Methods("POST")
	return router
}
