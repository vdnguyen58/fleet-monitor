package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vdnguyen58/fleet-monitor/handlers"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App) {
	// Initialize handlers
	deviceHandler := handlers.NewDeviceHandler()

	// API v1 group
	api := app.Group("/api/v1")

	// Device routes
	devices := api.Group("/devices")

	// POST /api/v1/devices/{device_id}/heartbeat
	devices.Post("/:device_id/heartbeat", deviceHandler.PostHeartbeat)

	// POST /api/v1/devices/{device_id}/stats
	devices.Post("/:device_id/stats", deviceHandler.PostStats)

	// GET /api/v1/devices/{device_id}/stats
	devices.Get("/:device_id/stats", deviceHandler.GetStats)
}
