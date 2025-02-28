package router

import (
	"github.com/Abhishekh669/backend/controller"
	"github.com/gorilla/mux"
)

func MessageRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/message", controller.CreateMessageHandler).Methods("POST")
	router.HandleFunc("/api/message/all/{id}", controller.GetAllConversationMessagesHandler).Methods("GET")
	return router
}
