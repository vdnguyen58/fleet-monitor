package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/vdnguyen58/fleet-monitor/routes"
)

func main() {
	// Create Fiber app with custom configuration
	app := fiber.New(fiber.Config{
		AppName:      "Fleet Management Metrics Server",
		ServerHeader: "Fiber",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New()) // Recover from panics
	app.Use(logger.New())  // Request logging
	app.Use(cors.New())    // CORS

	// Setup routes
	routes.SetupRoutes(app)

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "6733" // Default port from OpenAPI spec
	}

	log.Printf("Starting server on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}

// customErrorHandler handles errors returned from handlers
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Return error response
	return c.Status(code).JSON(fiber.Map{
		"msg": err.Error(),
	})
}
