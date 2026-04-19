package app

import (
	"context"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/algorithm"
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

	TableRepo    repository.TableRepo
	TableService services.TableService
	TableHandler handlers.TableHandler

	OrderRepo     repository.OrderRepo
	OrderRecCache *algorithm.CacheManager
	OrderService  services.OrderService
	Orderhandler  handlers.OrderHandler

	PaymentRepo    repository.PaymentRepo
	PaymentService services.PaymentService
	PaymentHandler handlers.PaymentHandler

	ReportRepo    repository.ReportRepo
	ReportCache   *algorithm.DefaultRevenueCache
	ReportHandler handlers.ReportHandler

	SettingRepo    repository.SettingRepo
	SettingService services.SettingService
	SettingHandler handlers.SettingHandler
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

	tableRepo := repository.NewTableRepository()
	tableService := services.NewTableService(tableRepo)
	tableHandler := handlers.NewTableHandler(tableService)

	orderRepo := repository.NewOrderRepository()
	orderRecCache := algorithm.NewCacheManager(nil, 48*time.Hour, orderRepo)
	orderService := services.NewOrderService(orderRepo, orderRecCache)
	orderHandler := handlers.NewOrderHandler(orderService)

	paymentRepo := repository.NewPaymentRepository()
	paymentService := services.NewPaymentService(paymentRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	reportRepo := repository.NewReportRepo()
	reportCache := algorithm.NewDefaultRevenueCache(reportRepo)
	reportHandler := handlers.NewReportHandler(reportRepo, reportCache) // pass same cache

	settingRepo := repository.NewSettingRepository()
	settingService := services.NewSettingService(settingRepo)
	settingHandler := handlers.NewSettingHandler(settingService)

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

		TableRepo:    tableRepo,
		TableService: tableService,
		TableHandler: *tableHandler,

		OrderRepo:     orderRepo,
		OrderRecCache: orderRecCache,
		OrderService:  orderService,
		Orderhandler:  *orderHandler,

		PaymentRepo:    paymentRepo,
		PaymentService: paymentService,
		PaymentHandler: *paymentHandler,

		ReportRepo:    reportRepo,
		ReportCache:   reportCache,
		ReportHandler: *reportHandler,

		SettingRepo:    settingRepo,
		SettingService: settingService,
		SettingHandler: *settingHandler,
	}, nil
}
