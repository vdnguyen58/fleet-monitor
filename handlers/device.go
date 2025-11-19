package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vdnguyen58/fleet-monitor/models"
)

// DeviceHandler handles device-related requests
type DeviceHandler struct {
	// Add dependencies here (e.g., database, services)
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{}
}

// PostHeartbeat handles POST /devices/{device_id}/heartbeat
func (h *DeviceHandler) PostHeartbeat(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")

	var req models.HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Msg: "Invalid request body",
		})
	}

	// TODO: Implement heartbeat logic
	// - Validate device exists
	// - Store heartbeat timestamp
	// - Update device last seen time

	// Example: Device not found
	// return c.Status(fiber.StatusNotFound).JSON(models.NotFoundResponse{
	// 	Msg: "Device not found",
	// })

	_ = deviceID // Remove this when implementing

	return c.SendStatus(fiber.StatusNoContent)
}

// PostStats handles POST /devices/{device_id}/stats
func (h *DeviceHandler) PostStats(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")

	var req models.UploadStatsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Msg: "Invalid request body",
		})
	}

	// TODO: Implement stats upload logic
	// - Validate device exists
	// - Store upload statistics
	// - Update aggregate metrics

	// Example: Device not found
	// return c.Status(fiber.StatusNotFound).JSON(models.NotFoundResponse{
	// 	Msg: "Device not found",
	// })

	_ = deviceID // Remove this when implementing

	return c.SendStatus(fiber.StatusNoContent)
}

// GetStats handles GET /devices/{device_id}/stats
func (h *DeviceHandler) GetStats(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")

	// TODO: Implement stats retrieval logic
	// - Validate device exists
	// - Calculate average upload time
	// - Calculate uptime percentage
	// - Return statistics

	// Example: Device not found
	// return c.Status(fiber.StatusNotFound).JSON(models.NotFoundResponse{
	// 	Msg: "Device not found",
	// })

	_ = deviceID // Remove this when implementing

	// Example response
	response := models.GetDeviceStatsResponse{
		AvgUploadTime: "5m10s",
		Uptime:        98.999,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
