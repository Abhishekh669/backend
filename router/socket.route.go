package router

import (
	"github.com/Abhishekh669/backend/sockets"
	"github.com/gorilla/mux"
)

func SocketRouter() *mux.Router {
	manager := sockets.GetManager()
	router := mux.NewRouter()
	router.HandleFunc("/api/ws", manager.ServeWs).Methods("GET")
	return router
}
