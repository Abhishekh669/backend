package router

import (
	"github.com/Abhishekh669/backend/controller"
	"github.com/gorilla/mux"
)

func ChatRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/chat", controller.CreateChatHandler).Methods("POST")
	router.HandleFunc("/api/chat/{id}", controller.GetConversationByIdHandler).Methods("GET")
	return router
}
