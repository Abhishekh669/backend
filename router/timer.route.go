package router

import (
	"github.com/Abhishekh669/backend/controller"
	"github.com/gorilla/mux"
)

func TimerRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/timer", controller.CreateTimerController).Methods("POST")
	router.HandleFunc("/api/timer/edit/{id}", controller.EditTimerDataHandler).Methods("POST")
	router.HandleFunc("/api/timer/delete/{id}", controller.DeleteTimerHandler).Methods("POST")
	router.HandleFunc("/api/timer/workspace", controller.CreateTimerWorkspaceHandler).Methods("POST")
	router.HandleFunc("/api/timer/workspace/delete/{id}", controller.DeleteWorkspaceHandler).Methods("DELETE")
	router.HandleFunc("/api/timer/workspace/edit/{id}", controller.EditTimerWorkspaceHandler).Methods("POST")
	router.HandleFunc("/api/timer/workspace", controller.GetUserTimerWorkspacesHandler).Methods("GET")
	router.HandleFunc("/api/timer/workspace/{id}", controller.GetUserWorkspaceTimersController).Methods("GET")
	return router
}
