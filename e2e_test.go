package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/vdnguyen58/fleet-monitor/models"
	"github.com/vdnguyen58/fleet-monitor/routes"
	"github.com/vdnguyen58/fleet-monitor/storage"
)

func setupE2EServer(t *testing.T) *fiber.App {
	// Create temporary CSV file with test devices
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
`
	tmpFile, err := os.CreateTemp("", "e2e-devices-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Initialize store
	store := storage.NewDeviceStore()
	if err := store.LoadDevicesFromCSV(tmpFile.Name()); err != nil {
		t.Fatalf("Failed to load devices: %v", err)
	}

	// Create app
	app := fiber.New(fiber.Config{
		AppName:      "Fleet Management Metrics Server",
		ServerHeader: "Fiber",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	routes.SetupRoutes(app, store)

	return app
}

// E2E Test 1: POST /api/v1/devices/{device_id}/heartbeat
func TestE2E_PostHeartbeat_Success(t *testing.T) {
	app := setupE2EServer(t)

	reqBody := models.HeartbeatRequest{
		SentAt: time.Now(),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/devices/60-6b-44-84-dc-64/heartbeat", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
}

func TestE2E_PostHeartbeat_DeviceNotFound(t *testing.T) {
	app := setupE2EServer(t)

	reqBody := models.HeartbeatRequest{
		SentAt: time.Now(),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/devices/invalid-device/heartbeat", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var notFoundResp models.NotFoundResponse
	json.Unmarshal(body, &notFoundResp)

	if notFoundResp.Msg != "Device not found" {
		t.Errorf("Expected 'Device not found', got '%s'", notFoundResp.Msg)
	}
}

func TestE2E_PostHeartbeat_InvalidJSON(t *testing.T) {
	app := setupE2EServer(t)

	req := httptest.NewRequest("POST", "/api/v1/devices/60-6b-44-84-dc-64/heartbeat", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

// E2E Test 2: POST /api/v1/devices/{device_id}/stats
func TestE2E_PostStats_Success(t *testing.T) {
	app := setupE2EServer(t)

	reqBody := models.UploadStatsRequest{
		SentAt:     time.Now(),
		UploadTime: 5000000000, // 5 seconds in nanoseconds
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/devices/b4-45-52-a2-f1-3c/stats", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
}

func TestE2E_PostStats_DeviceNotFound(t *testing.T) {
	app := setupE2EServer(t)

	reqBody := models.UploadStatsRequest{
		SentAt:     time.Now(),
		UploadTime: 5000000000,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/devices/invalid-device/stats", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestE2E_PostStats_InvalidJSON(t *testing.T) {
	app := setupE2EServer(t)

	req := httptest.NewRequest("POST", "/api/v1/devices/b4-45-52-a2-f1-3c/stats", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

// E2E Test 3: GET /api/v1/devices/{device_id}/stats
func TestE2E_GetStats_Success(t *testing.T) {
	app := setupE2EServer(t)
	deviceID := "26-9a-66-01-33-83"

	// First, send some heartbeats
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		heartbeat := models.HeartbeatRequest{
			SentAt: baseTime.Add(time.Duration(i) * time.Minute),
		}
		bodyBytes, _ := json.Marshal(heartbeat)
		req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/heartbeat", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req)
	}

	// Send some upload stats
	uploadTimes := []int64{3000000000, 6000000000, 9000000000} // 3s, 6s, 9s (avg = 6s)
	for _, uploadTime := range uploadTimes {
		stats := models.UploadStatsRequest{
			SentAt:     time.Now(),
			UploadTime: uploadTime,
		}
		bodyBytes, _ := json.Marshal(stats)
		req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/stats", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req)
	}

	// Now get the stats
	req := httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var statsResp models.GetDeviceStatsResponse
	if err := json.Unmarshal(body, &statsResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify avg_upload_time
	if statsResp.AvgUploadTime != "6s" {
		t.Errorf("Expected avg_upload_time '6s', got '%s'", statsResp.AvgUploadTime)
	}

	// Verify uptime (5 heartbeats over 4 minutes = 125%)
	expectedUptime := (5.0 / 4.0) * 100.0
	if statsResp.Uptime != expectedUptime {
		t.Errorf("Expected uptime %f, got %f", expectedUptime, statsResp.Uptime)
	}
}

func TestE2E_GetStats_NoData(t *testing.T) {
	app := setupE2EServer(t)

	// Request stats for device with no data
	req := httptest.NewRequest("GET", "/api/v1/devices/60-6b-44-84-dc-64/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204 (no content), got %d", resp.StatusCode)
	}
}

func TestE2E_GetStats_DeviceNotFound(t *testing.T) {
	app := setupE2EServer(t)

	req := httptest.NewRequest("GET", "/api/v1/devices/invalid-device/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestE2E_GetStats_OnlyHeartbeats(t *testing.T) {
	app := setupE2EServer(t)
	deviceID := "60-6b-44-84-dc-64"

	// Send only heartbeats
	heartbeat := models.HeartbeatRequest{
		SentAt: time.Now(),
	}
	bodyBytes, _ := json.Marshal(heartbeat)
	req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/heartbeat", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)

	// Get stats
	req = httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var statsResp models.GetDeviceStatsResponse
	json.Unmarshal(body, &statsResp)

	// Should have 100% uptime (single heartbeat)
	if statsResp.Uptime != 100.0 {
		t.Errorf("Expected uptime 100.0, got %f", statsResp.Uptime)
	}

	// Should have "0s" for avg_upload_time (no upload data)
	if statsResp.AvgUploadTime != "0s" {
		t.Errorf("Expected avg_upload_time '0s', got '%s'", statsResp.AvgUploadTime)
	}
}

func TestE2E_GetStats_OnlyUploadTimes(t *testing.T) {
	app := setupE2EServer(t)
	deviceID := "b4-45-52-a2-f1-3c"

	// Send only upload stats
	stats := models.UploadStatsRequest{
		SentAt:     time.Now(),
		UploadTime: 10000000000, // 10 seconds
	}
	bodyBytes, _ := json.Marshal(stats)
	req := httptest.NewRequest("POST", "/api/v1/devices/"+deviceID+"/stats", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)

	// Get stats
	req = httptest.NewRequest("GET", "/api/v1/devices/"+deviceID+"/stats", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var statsResp models.GetDeviceStatsResponse
	json.Unmarshal(body, &statsResp)

	// Should have 0% uptime (no heartbeats)
	if statsResp.Uptime != 0.0 {
		t.Errorf("Expected uptime 0.0, got %f", statsResp.Uptime)
	}

	// Should have "10s" for avg_upload_time
	if statsResp.AvgUploadTime != "10s" {
		t.Errorf("Expected avg_upload_time '10s', got '%s'", statsResp.AvgUploadTime)
	}
}
