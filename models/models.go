package models

import "time"

// HeartbeatRequest represents a heartbeat from a device
type HeartbeatRequest struct {
	SentAt time.Time `json:"sent_at" validate:"required"`
}

// UploadStatsRequest represents device statistics
type UploadStatsRequest struct {
	SentAt     time.Time `json:"sent_at" validate:"required"`
	UploadTime int64     `json:"upload_time" validate:"required"` // nanoseconds
}

// GetDeviceStatsResponse represents device statistics response
type GetDeviceStatsResponse struct {
	AvgUploadTime string  `json:"avg_upload_time"` // duration string like "5m10s"
	Uptime        float64 `json:"uptime"`          // percentage like 98.999
}

// ErrorResponse represents a server error
type ErrorResponse struct {
	Msg string `json:"msg"`
}

// NotFoundResponse represents a not found error
type NotFoundResponse struct {
	Msg string `json:"msg"`
}
