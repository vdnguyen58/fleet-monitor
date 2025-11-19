package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/vdnguyen58/fleet-monitor/models"
)

const (
	baseURL    = "http://localhost:6733"
	apiV1      = baseURL + "/api/v1"
	maxRetries = 10
	retryDelay = 500 * time.Millisecond
)

// waitForServer waits for the server to be ready
func waitForServer(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			t.Log("Server is ready")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(retryDelay)
	}
	t.Fatal("Server did not become ready in time")
}

// Docker E2E Test 1: POST /api/v1/devices/{device_id}/heartbeat
func TestDocker_PostHeartbeat_Success(t *testing.T) {
	waitForServer(t)

	reqBody := models.HeartbeatRequest{
		SentAt: time.Now(),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		apiV1+"/devices/60-6b-44-84-dc-64/heartbeat",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
}

func TestDocker_PostHeartbeat_DeviceNotFound(t *testing.T) {
	waitForServer(t)

	reqBody := models.HeartbeatRequest{
		SentAt: time.Now(),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		apiV1+"/devices/invalid-device/heartbeat",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var notFoundResp models.NotFoundResponse
	if err := json.Unmarshal(body, &notFoundResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if notFoundResp.Msg != "Device not found" {
		t.Errorf("Expected 'Device not found', got '%s'", notFoundResp.Msg)
	}
}

// Docker E2E Test 2: POST /api/v1/devices/{device_id}/stats
func TestDocker_PostStats_Success(t *testing.T) {
	waitForServer(t)

	reqBody := models.UploadStatsRequest{
		SentAt:     time.Now(),
		UploadTime: 5000000000, // 5 seconds
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		apiV1+"/devices/b4-45-52-a2-f1-3c/stats",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
}

func TestDocker_PostStats_DeviceNotFound(t *testing.T) {
	waitForServer(t)

	reqBody := models.UploadStatsRequest{
		SentAt:     time.Now(),
		UploadTime: 5000000000,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		apiV1+"/devices/invalid-device/stats",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

// Docker E2E Test 3: GET /api/v1/devices/{device_id}/stats
func TestDocker_GetStats_WithData(t *testing.T) {
	waitForServer(t)
	deviceID := "26-9a-66-01-33-83"

	// Send heartbeats
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		heartbeat := models.HeartbeatRequest{
			SentAt: baseTime.Add(time.Duration(i) * time.Minute),
		}
		bodyBytes, _ := json.Marshal(heartbeat)

		resp, err := http.Post(
			fmt.Sprintf("%s/devices/%s/heartbeat", apiV1, deviceID),
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Failed to send heartbeat: %v", err)
		}
		resp.Body.Close()
	}

	// Send upload stats
	uploadTimes := []int64{3000000000, 6000000000, 9000000000} // 3s, 6s, 9s (avg = 6s)
	for _, uploadTime := range uploadTimes {
		stats := models.UploadStatsRequest{
			SentAt:     time.Now(),
			UploadTime: uploadTime,
		}
		bodyBytes, _ := json.Marshal(stats)

		resp, err := http.Post(
			fmt.Sprintf("%s/devices/%s/stats", apiV1, deviceID),
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Failed to send stats: %v", err)
		}
		resp.Body.Close()
	}

	// Get stats
	resp, err := http.Get(fmt.Sprintf("%s/devices/%s/stats", apiV1, deviceID))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

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

func TestDocker_GetStats_NoData(t *testing.T) {
	waitForServer(t)

	// Use a device with no data yet
	resp, err := http.Get(apiV1 + "/devices/18-b8-87-e7-1f-06/stats")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("Expected status 204 (no content), got %d", resp.StatusCode)
	}
}

func TestDocker_GetStats_DeviceNotFound(t *testing.T) {
	waitForServer(t)

	resp, err := http.Get(apiV1 + "/devices/invalid-device/stats")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestDocker_HealthCheck(t *testing.T) {
	waitForServer(t)

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var healthResp map[string]string
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if healthResp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", healthResp["status"])
	}
}
