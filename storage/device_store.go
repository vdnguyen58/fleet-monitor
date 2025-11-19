package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

// DeviceData holds the tracking data for a single device
type DeviceData struct {
	Heartbeats  []time.Time // timestamps of heartbeats
	UploadTimes []int64     // upload times in nanoseconds
	mu          sync.RWMutex
}

// DeviceStore manages all device data
type DeviceStore struct {
	devices map[string]*DeviceData
	mu      sync.RWMutex
}

// NewDeviceStore creates a new device store
func NewDeviceStore() *DeviceStore {
	return &DeviceStore{
		devices: make(map[string]*DeviceData),
	}
}

// LoadDevicesFromCSV loads device IDs from a CSV file
func (s *DeviceStore) LoadDevicesFromCSV(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	// Skip header row and load device IDs
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, record := range records {
		if i == 0 {
			// Skip header
			continue
		}
		if len(record) > 0 && record[0] != "" {
			deviceID := record[0]
			s.devices[deviceID] = &DeviceData{
				Heartbeats:  make([]time.Time, 0),
				UploadTimes: make([]int64, 0),
			}
		}
	}

	return nil
}

// DeviceExists checks if a device ID exists
func (s *DeviceStore) DeviceExists(deviceID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.devices[deviceID]
	return exists
}

// AddHeartbeat adds a heartbeat timestamp for a device
func (s *DeviceStore) AddHeartbeat(deviceID string, timestamp time.Time) error {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("device not found")
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	device.Heartbeats = append(device.Heartbeats, timestamp)
	return nil
}

// AddUploadTime adds an upload time for a device
func (s *DeviceStore) AddUploadTime(deviceID string, uploadTime int64) error {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("device not found")
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	device.UploadTimes = append(device.UploadTimes, uploadTime)
	return nil
}

// GetDeviceData retrieves a copy of device data
func (s *DeviceStore) GetDeviceData(deviceID string) (*DeviceData, error) {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device not found")
	}

	device.mu.RLock()
	defer device.mu.RUnlock()

	// Return a copy to avoid race conditions
	copy := &DeviceData{
		Heartbeats:  make([]time.Time, len(device.Heartbeats)),
		UploadTimes: make([]int64, len(device.UploadTimes)),
	}
	copySlice(copy.Heartbeats, device.Heartbeats)
	copyInt64Slice(copy.UploadTimes, device.UploadTimes)

	return copy, nil
}

// Helper function to copy time slices
func copySlice(dst, src []time.Time) {
	for i, v := range src {
		dst[i] = v
	}
}

// Helper function to copy int64 slices
func copyInt64Slice(dst, src []int64) {
	for i, v := range src {
		dst[i] = v
	}
}
