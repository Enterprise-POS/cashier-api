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
		log.SetFormatter(&log.TextFormatter{})
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
		return ctx.Status(fiber.StatusOK).
			JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
				"message": "Welcome to API Gateway",
			}))
	})

	// public
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := controller.NewUserControllerImpl(userService)

	apiV1.Post("/users/sign_up", userController.SignUpWithEmailAndPassword)
	apiV1.Post("/users/sign_in", userController.SignInWithEmailAndPassword)
	apiV1.Delete("/users/sign_out", userController.SignOut)

	// protected only login user
	apiV1.Use(middleware.ProtectedRoute)

	tenantRepository := repository.NewTenantRepositoryImpl(supabaseClient)
	tenantService := service.NewTenantServiceImpl(tenantRepository)
	tenantController := controller.NewTenantControllerImpl(tenantService)

	apiV1.Get("/tenants/:userId", tenantController.GetTenantWithUser)
	apiV1.Get("/tenants/members/:tenantId", tenantController.GetTenantMembers)
	apiV1.Post("/tenants/new", tenantController.NewTenant)
	apiV1.Post("/tenants/add_user", tenantController.AddUserToTenant)
	apiV1.Delete("/tenants/remove_user", tenantController.RemoveUserFromTenant)

	// restrict by tenantId
	tenantRestriction := middleware.RestrictByTenant(supabaseClient)

	warehouseRepository := repository.NewWarehouseRepositoryImpl(supabaseClient)
	warehouseService := service.NewWarehouseServiceImpl(warehouseRepository)
	warehouseController := controller.NewWarehouseControllerImpl(warehouseService)

	// GET /warehouses/:tenantId?limit=10&page=1
	apiV1.Get("/warehouses/:tenantId", tenantRestriction, warehouseController.Get)
	apiV1.Get("/warehouses/active/:tenantId", tenantRestriction, warehouseController.GetActiveItem)
	apiV1.Post("/warehouses/create_item/:tenantId", tenantRestriction, warehouseController.CreateItem)
	apiV1.Post("/warehouses/find/:tenantId", tenantRestriction, warehouseController.FindById)
	apiV1.Post("/warehouses/find_complete_by_id/:tenantId", tenantRestriction, warehouseController.FindCompleteById)
	apiV1.Put("/warehouses/edit/:tenantId", tenantRestriction, warehouseController.Edit)
	apiV1.Put("/warehouses/activate/:tenantId", tenantRestriction, warehouseController.SetActivate)

	categoryRepository := repository.NewCategoryRepositoryImpl(supabaseClient)
	categoryService := service.NewCategoryServiceImpl(categoryRepository)
	categoryController := controller.NewCategoryControllerImpl(categoryService)

	// GET /categories/:tenantId?limit=10&page=1
	apiV1.Get("/categories/:tenantId", tenantRestriction, categoryController.Get)
	apiV1.Post("/categories/items_by_category_id/:tenantId", tenantRestriction, categoryController.GetItemsByCategoryId)
	apiV1.Post("/categories/category_with_items/:tenantId", tenantRestriction, categoryController.GetCategoryWithItems)
	apiV1.Post("/categories/create/:tenantId", tenantRestriction, categoryController.Create)
	apiV1.Post("/categories/register/:tenantId", tenantRestriction, categoryController.Register)
	apiV1.Put("/categories/update/:tenantId", tenantRestriction, categoryController.Update)
	apiV1.Put("/categories/edit_item_category/:tenantId", tenantRestriction, categoryController.EditItemCategory)
	apiV1.Delete("/categories/unregister/:tenantId", tenantRestriction, categoryController.Unregister)
	apiV1.Delete("/categories/:tenantId", tenantRestriction, categoryController.Delete)

	// Handle route not found (404)
	app.All("*", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusNotFound).
			JSON(common.NewWebResponseError(fiber.StatusNotFound, common.StatusError, "Route for "+string(ctx.Request().RequestURI())+" not found"))
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
