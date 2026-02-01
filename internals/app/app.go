package app

import (
	"context"
	"log"

	"github.com/Abhishekh669/backend/internals/handlers"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/Abhishekh669/backend/internals/services"
)

type App struct {
	UserRepo    repository.UserRepo
	UserService services.UserService
	UserHandler handlers.UserHandler

	RawMaterialRepo    repository.RawMaterialsRepo
	RawMaterialService services.RawMaterialService
	RawMaterialHandler handlers.RawMaterialsHandler
}

func New() (*App, error) {
	userRepo := repository.NewUserRepository()

	err := userRepo.EnsureAdminUserExists(context.Background())
	if err != nil {
		log.Println("error in ensuring admin user exists: ", err)
		return nil, err
	}

	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	rawMaterialsRepo := repository.NewRawMaterialsRepository()
	rawMaterialsService := services.NewRawMaterialService(rawMaterialsRepo)
	rawMaterialsHandler := handlers.NewRawMaterialHandler(rawMaterialsService)

	return &App{
		UserRepo:    userRepo,
		UserService: userService,
		UserHandler: *userHandler,

		RawMaterialRepo:    rawMaterialsRepo,
		RawMaterialService: rawMaterialsService,
		RawMaterialHandler: *rawMaterialsHandler,
	}, nil
}
