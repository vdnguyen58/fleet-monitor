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

type testCase struct {
	name           string                        // Test description
	method         string                        // HTTP method (GET, POST, etc.)
	path           string                        // URL path
	body           interface{}                   // Request body (will be JSON marshaled)
	expectedStatus int                           // Expected HTTP status code
	setup          func(t *testing.T)            // Optional setup before test
	validate       func(t *testing.T, body []byte) // Optional custom validation
}

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

// sendRequest sends an HTTP request and returns the response
func sendRequest(t *testing.T, method, url string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, responseBody
}

// runTestCase executes a single test case
func runTestCase(t *testing.T, tc testCase) {
	t.Helper()

	// Run setup if provided
	if tc.setup != nil {
		tc.setup(t)
	}

	// Send request
	resp, body := sendRequest(t, tc.method, tc.path, tc.body)

	// Validate status code
	if resp.StatusCode != tc.expectedStatus {
		t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
		if len(body) > 0 {
			t.Logf("Response body: %s", string(body))
		}
	}

	// Run custom validation if provided
	if tc.validate != nil {
		tc.validate(t, body)
	}
}

// Helper function to send heartbeats for a device
func sendHeartbeats(t *testing.T, deviceID string, count int, startTime time.Time) {
	t.Helper()
	for i := 0; i < count; i++ {
		heartbeat := models.HeartbeatRequest{
			SentAt: startTime.Add(time.Duration(i) * time.Minute),
		}
		url := fmt.Sprintf("%s/devices/%s/heartbeat", apiV1, deviceID)
		resp, _ := sendRequest(t, "POST", url, heartbeat)
		if resp.StatusCode != 204 {
			t.Fatalf("Failed to send heartbeat: status %d", resp.StatusCode)
		}
	}
}

// Helper function to send upload stats for a device
func sendUploadStats(t *testing.T, deviceID string, uploadTimes []int64) {
	t.Helper()
	for _, uploadTime := range uploadTimes {
		stats := models.UploadStatsRequest{
			SentAt:     time.Now(),
			UploadTime: uploadTime,
		}
		url := fmt.Sprintf("%s/devices/%s/stats", apiV1, deviceID)
		resp, _ := sendRequest(t, "POST", url, stats)
		if resp.StatusCode != 204 {
			t.Fatalf("Failed to send stats: status %d", resp.StatusCode)
		}
	}
}

// Heartbeat API
func TestDocker_Heartbeat(t *testing.T) {
	waitForServer(t)

	testCases := []testCase{
		{
			name:           "Successful heartbeat",
			method:         "POST",
			path:           apiV1 + "/devices/60-6b-44-84-dc-64/heartbeat",
			body:           models.HeartbeatRequest{SentAt: time.Now()},
			expectedStatus: 204,
		},
		{
			name:           "Device not found",
			method:         "POST",
			path:           apiV1 + "/devices/invalid-device/heartbeat",
			body:           models.HeartbeatRequest{SentAt: time.Now()},
			expectedStatus: 404,
			validate: func(t *testing.T, body []byte) {
				var notFoundResp models.NotFoundResponse
				if err := json.Unmarshal(body, &notFoundResp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if notFoundResp.Msg != "Device not found" {
					t.Errorf("Expected 'Device not found', got '%s'", notFoundResp.Msg)
				}
			},
		},
		{
			name:           "Invalid JSON",
			method:         "POST",
			path:           apiV1 + "/devices/60-6b-44-84-dc-64/heartbeat",
			body:           nil, // Will send empty body
			expectedStatus: 400,
		},
		{
			name:           "Multiple heartbeats for same device",
			method:         "POST",
			path:           apiV1 + "/devices/b4-45-52-a2-f1-3c/heartbeat",
			body:           models.HeartbeatRequest{SentAt: time.Now()},
			expectedStatus: 204,
			setup: func(t *testing.T) {
				// Send initial heartbeat
				sendHeartbeats(t, "b4-45-52-a2-f1-3c", 1, time.Now().Add(-1*time.Minute))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

// Upload Stats API
func TestDocker_UploadStats(t *testing.T) {
	waitForServer(t)

	testCases := []testCase{
		{
			name:   "Successful stats upload",
			method: "POST",
			path:   apiV1 + "/devices/b4-45-52-a2-f1-3c/stats",
			body: models.UploadStatsRequest{
				SentAt:     time.Now(),
				UploadTime: 5000000000, // 5 seconds
			},
			expectedStatus: 204,
		},
		{
			name:   "Device not found",
			method: "POST",
			path:   apiV1 + "/devices/invalid-device/stats",
			body: models.UploadStatsRequest{
				SentAt:     time.Now(),
				UploadTime: 5000000000,
			},
			expectedStatus: 404,
		},
		{
			name:           "Invalid JSON",
			method:         "POST",
			path:           apiV1 + "/devices/b4-45-52-a2-f1-3c/stats",
			body:           nil,
			expectedStatus: 400,
		},
		{
			name:   "Large upload time",
			method: "POST",
			path:   apiV1 + "/devices/26-9a-66-01-33-83/stats",
			body: models.UploadStatsRequest{
				SentAt:     time.Now(),
				UploadTime: 300000000000, // 5 minutes
			},
			expectedStatus: 204,
		},
		{
			name:   "Small upload time (milliseconds)",
			method: "POST",
			path:   apiV1 + "/devices/26-9a-66-01-33-83/stats",
			body: models.UploadStatsRequest{
				SentAt:     time.Now(),
				UploadTime: 100000000, // 100ms
			},
			expectedStatus: 204,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

// Get Stats API
func TestDocker_GetStats(t *testing.T) {
	waitForServer(t)

	testCases := []testCase{
		{
			name:           "Get stats with heartbeats and upload times",
			method:         "GET",
			path:           apiV1 + "/devices/26-9a-66-01-33-83/stats",
			expectedStatus: 200,
			setup: func(t *testing.T) {
				// Send 5 heartbeats (1 per minute)
				sendHeartbeats(t, "26-9a-66-01-33-83", 5, time.Now())
				// Send upload stats: avg = 6s
				sendUploadStats(t, "26-9a-66-01-33-83", []int64{
					3000000000,  // 3s
					6000000000,  // 6s
					9000000000,  // 9s
				})
			},
			validate: func(t *testing.T, body []byte) {
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
			},
		},
		{
			name:           "Get stats with no data",
			method:         "GET",
			path:           apiV1 + "/devices/18-b8-87-e7-1f-06/stats",
			expectedStatus: 204,
		},
		{
			name:           "Get stats for non-existent device",
			method:         "GET",
			path:           apiV1 + "/devices/invalid-device/stats",
			expectedStatus: 404,
		},
		{
			name:           "Get stats with only heartbeats",
			method:         "GET",
			path:           apiV1 + "/devices/60-6b-44-84-dc-64/stats",
			expectedStatus: 200,
			setup: func(t *testing.T) {
				sendHeartbeats(t, "60-6b-44-84-dc-64", 1, time.Now())
			},
			validate: func(t *testing.T, body []byte) {
				var statsResp models.GetDeviceStatsResponse
				if err := json.Unmarshal(body, &statsResp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Single heartbeat = 100% uptime
				if statsResp.Uptime != 100.0 {
					t.Errorf("Expected uptime 100.0, got %f", statsResp.Uptime)
				}

				// No upload data = "0s"
				if statsResp.AvgUploadTime != "0s" {
					t.Errorf("Expected avg_upload_time '0s', got '%s'", statsResp.AvgUploadTime)
				}
			},
		},
		{
			name:           "Get stats with only upload times",
			method:         "GET",
			path:           apiV1 + "/devices/38-4e-73-e0-33-59/stats",
			expectedStatus: 200,
			setup: func(t *testing.T) {
				sendUploadStats(t, "38-4e-73-e0-33-59", []int64{10000000000}) // 10s
			},
			validate: func(t *testing.T, body []byte) {
				var statsResp models.GetDeviceStatsResponse
				if err := json.Unmarshal(body, &statsResp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// No heartbeats = 0% uptime
				if statsResp.Uptime != 0.0 {
					t.Errorf("Expected uptime 0.0, got %f", statsResp.Uptime)
				}

				// Average upload time = 10s
				if statsResp.AvgUploadTime != "10s" {
					t.Errorf("Expected avg_upload_time '10s', got '%s'", statsResp.AvgUploadTime)
				}
			},
		},
		{
			name:           "Get stats with missed heartbeats",
			method:         "GET",
			path:           apiV1 + "/devices/b4-45-52-a2-f1-3c/stats",
			expectedStatus: 200,
			setup: func(t *testing.T) {
				// 3 heartbeats over 5 minutes (2 missed) = 60% uptime
				baseTime := time.Now()
				heartbeats := []models.HeartbeatRequest{
					{SentAt: baseTime},
					{SentAt: baseTime.Add(1 * time.Minute)},
					{SentAt: baseTime.Add(5 * time.Minute)},
				}
				for _, hb := range heartbeats {
					url := fmt.Sprintf("%s/devices/b4-45-52-a2-f1-3c/heartbeat", apiV1)
					sendRequest(t, "POST", url, hb)
				}
			},
			validate: func(t *testing.T, body []byte) {
				var statsResp models.GetDeviceStatsResponse
				if err := json.Unmarshal(body, &statsResp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// 3 heartbeats over 5 minutes = 60% uptime
				expectedUptime := (3.0 / 5.0) * 100.0
				if statsResp.Uptime != expectedUptime {
					t.Errorf("Expected uptime %f, got %f", expectedUptime, statsResp.Uptime)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

// Test health check endpoint
func TestDocker_HealthCheck(t *testing.T) {
	waitForServer(t)

	testCases := []testCase{
		{
			name:           "Health check returns OK",
			method:         "GET",
			path:           baseURL + "/health",
			expectedStatus: 200,
			validate: func(t *testing.T, body []byte) {
				var healthResp map[string]string
				if err := json.Unmarshal(body, &healthResp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if healthResp["status"] != "ok" {
					t.Errorf("Expected status 'ok', got '%s'", healthResp["status"])
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}
