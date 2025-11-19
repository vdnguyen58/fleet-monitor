package handlers

import (
	"fmt"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vdnguyen58/fleet-monitor/models"
	"github.com/vdnguyen58/fleet-monitor/storage"
)

// DeviceHandler handles device-related requests
type DeviceHandler struct {
	store *storage.DeviceStore
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(store *storage.DeviceStore) *DeviceHandler {
	return &DeviceHandler{
		store: store,
	}
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

	// Validate device exists
	if !h.store.DeviceExists(deviceID) {
		return c.Status(fiber.StatusNotFound).JSON(models.NotFoundResponse{
			Msg: "Device not found",
		})
	}

	// Store heartbeat timestamp
	if err := h.store.AddHeartbeat(deviceID, req.SentAt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Msg: fmt.Sprintf("Failed to store heartbeat: %v", err),
		})
	}

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

	// Validate device exists
	if !h.store.DeviceExists(deviceID) {
		return c.Status(fiber.StatusNotFound).JSON(models.NotFoundResponse{
			Msg: "Device not found",
		})
	}

	// Store upload time
	if err := h.store.AddUploadTime(deviceID, req.UploadTime); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Msg: fmt.Sprintf("Failed to store upload time: %v", err),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetStats handles GET /devices/{device_id}/stats
func (h *DeviceHandler) GetStats(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")

	// Validate device exists
	if !h.store.DeviceExists(deviceID) {
		return c.Status(fiber.StatusNotFound).JSON(models.NotFoundResponse{
			Msg: "Device not found",
		})
	}

	// Get device data
	deviceData, err := h.store.GetDeviceData(deviceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Msg: fmt.Sprintf("Failed to retrieve device data: %v", err),
		})
	}

	// If no data available yet, return 204
	if len(deviceData.Heartbeats) == 0 && len(deviceData.UploadTimes) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	// Calculate uptime
	uptime := calculateUptime(deviceData.Heartbeats)

	// Calculate average upload time
	avgUploadTime := calculateAvgUploadTime(deviceData.UploadTimes)

	response := models.GetDeviceStatsResponse{
		AvgUploadTime: avgUploadTime,
		Uptime:        uptime,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// calculateUptime calculates device uptime percentage
// uptime = (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
func calculateUptime(heartbeats []time.Time) float64 {
	if len(heartbeats) == 0 {
		return 0.0
	}

	// Single heartbeat means 100% uptime
	if len(heartbeats) == 1 {
		return 100.0
	}

	// Sort heartbeats to find first and last
	sortedHeartbeats := make([]time.Time, len(heartbeats))
	copy(sortedHeartbeats, heartbeats)
	sort.Slice(sortedHeartbeats, func(i, j int) bool {
		return sortedHeartbeats[i].Before(sortedHeartbeats[j])
	})

	first := sortedHeartbeats[0]
	last := sortedHeartbeats[len(sortedHeartbeats)-1]

	// Calculate minutes between first and last heartbeat
	duration := last.Sub(first)
	minutes := duration.Minutes()

	// If duration is less than a minute, consider it 100% uptime
	if minutes < 1.0 {
		return 100.0
	}

	// Calculate uptime percentage
	numHeartbeats := float64(len(heartbeats))
	uptime := (numHeartbeats / minutes) * 100.0

	return uptime
}

// calculateAvgUploadTime calculates average upload time and formats as duration string
func calculateAvgUploadTime(uploadTimes []int64) string {
	if len(uploadTimes) == 0 {
		return "0s"
	}

	// Calculate average
	var sum int64
	for _, t := range uploadTimes {
		sum += t
	}
	avgNanos := sum / int64(len(uploadTimes))

	// Convert to time.Duration and format
	duration := time.Duration(avgNanos)
	return duration.String()
}
