package main

import (
	"cashier-api/controller"
	"cashier-api/exception"
	common "cashier-api/helper"
	"cashier-api/helper/client"
	"cashier-api/middleware"
	"cashier-api/repository"
	"cashier-api/service"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	if os.Getenv("MODE") == "prod" {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetFormatter(&log.TextFormatter{})
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	// 01 Set up the fiber and database
	app := fiber.New(fiber.Config{
		IdleTimeout:             time.Second * 5,
		ReadTimeout:             time.Second * 5,
		WriteTimeout:            time.Second * 5,
		Prefork:                 false,
		EnableTrustedProxyCheck: true,
		ErrorHandler:            exception.ErrorHandler,
	})

	// DB client
	supabaseClient := client.CreateSupabaseClient()

	// 02 Middleware, Security
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins:     "http://localhost:3000",
	// 	AllowCredentials: true,
	// }))
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))

	// 03 Router (grouping by /api/v1)
	apiV1 := app.Group("/api/v1")

	apiV1.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(200).JSON(common.WebResponse{
			Code:   200,
			Status: "OK",
			Data: fiber.Map{
				"message": "Welcome to API Gateway",
			},
		})
	})

	tenantRepository := repository.NewTenantRepositoryImpl(supabaseClient)
	tenantService := service.NewTenantServiceImpl(tenantRepository)
	tenantController := controller.NewTenantControllerImpl(tenantService)

	apiV1.Get("/tenants/:userId", middleware.ProtectedRoute, tenantController.GetTenantWithUser)
	apiV1.Get("/tenants/members/:tenantId", middleware.ProtectedRoute, tenantController.GetTenantMembers)
	apiV1.Post("/tenants/new", middleware.ProtectedRoute, tenantController.NewTenant)
	apiV1.Post("/tenants/add_user", middleware.ProtectedRoute, tenantController.AddUserToTenant)
	apiV1.Delete("/tenants/remove_user", middleware.ProtectedRoute, tenantController.RemoveUserFromTenant)

	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := controller.NewUserControllerImpl(userService)

	apiV1.Post("/users/sign_up", userController.SignUpWithEmailAndPassword)
	apiV1.Post("/users/sign_in", userController.SignInWithEmailAndPassword)
	apiV1.Delete("/users/sign_out", userController.SignOut)

	warehouseRepository := repository.NewWarehouseRepositoryImpl(supabaseClient)
	warehouseService := service.NewWarehouseServiceImpl(warehouseRepository)
	warehouseController := controller.NewWarehouseControllerImpl(warehouseService)

	// GET /warehouse/:tenantId?limit=10&page=1
	apiV1.Get("/warehouses/:id", middleware.ProtectedRoute, warehouseController.Get)

	// Handle route not found (404)
	app.All("*", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusNotFound).JSON(common.WebResponse{
			Code:   404,
			Status: "Not Found",
			Data: fiber.Map{
				"message": "Route for " + string(ctx.Request().RequestURI()) + " not found",
			},
		})
	})

	// 04 Application started listening here
	url := "localhost:8000"
	if os.Getenv("MODE") == "prod" {
		url = ":8000"
	}

	// The application began to listen to HTTP request
	log.Info("Start listening at " + url)
	err := app.Listen(url)
	if err != nil {
		panic(err)
	}
}
