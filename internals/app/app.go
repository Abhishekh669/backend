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

	FoodCategoryRepo    repository.FoodCategoryRepo
	FoodCategoryService services.FoodCategoryService
	FoodCategoryHandler handlers.FoodCategoryHandler

	AttendanceRepo    repository.AttendanceRepo
	AttendanceService services.AttendanceService
	AttendanceHandler handlers.AttendanceHandler
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

	foodCategoryRepo := repository.NewFoodCategoryRepository()
	foodCategoryService := services.NewFoodCategoryService(foodCategoryRepo)
	foodCategoryHandler := handlers.NewFoodCategoryHandler(foodCategoryService)

	attendanceRepo := repository.NewAttendanceRepository()
	attendanceService := services.NewAttendanceService(attendanceRepo)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceService)

	return &App{
		UserRepo:    userRepo,
		UserService: userService,
		UserHandler: *userHandler,

		RawMaterialRepo:    rawMaterialsRepo,
		RawMaterialService: rawMaterialsService,
		RawMaterialHandler: *rawMaterialsHandler,

		FoodCategoryRepo:    foodCategoryRepo,
		FoodCategoryService: foodCategoryService,
		FoodCategoryHandler: *foodCategoryHandler,

		AttendanceRepo:    attendanceRepo,
		AttendanceService: attendanceService,
		AttendanceHandler: *attendanceHandler,
	}, nil
}
