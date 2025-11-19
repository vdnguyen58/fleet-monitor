package storage

import (
	"os"
	"testing"
)

func TestLoadDevicesFromCSV_Success(t *testing.T) {
	csvContent := `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
`
	tmpFile, err := os.CreateTemp("", "devices-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	store := NewDeviceStore()
	err = store.LoadDevicesFromCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadDevicesFromCSV failed: %v", err)
	}

	// Verify devices were loaded
	expectedDevices := []string{
		"60-6b-44-84-dc-64",
		"b4-45-52-a2-f1-3c",
		"26-9a-66-01-33-83",
	}

	for _, deviceID := range expectedDevices {
		if !store.DeviceExists(deviceID) {
			t.Errorf("Expected device %s to exist", deviceID)
		}
	}
}

func TestLoadDevicesFromCSV_FileNotFound(t *testing.T) {
	store := NewDeviceStore()
	err := store.LoadDevicesFromCSV("nonexistent-file.csv")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestLoadDevicesFromCSV_EmptyFile(t *testing.T) {
	csvContent := `device_id
`
	tmpFile, err := os.CreateTemp("", "devices-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	store := NewDeviceStore()
	err = store.LoadDevicesFromCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadDevicesFromCSV failed: %v", err)
	}

	// Should have no devices
	if store.DeviceExists("") {
		t.Error("Should not have empty device ID")
	}
}

func TestLoadDevicesFromCSV_WithEmptyLines(t *testing.T) {
	csvContent := `device_id
device-1

device-2

`
	tmpFile, err := os.CreateTemp("", "devices-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	store := NewDeviceStore()
	err = store.LoadDevicesFromCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadDevicesFromCSV failed: %v", err)
	}

	// Should only load non-empty device IDs
	if !store.DeviceExists("device-1") {
		t.Error("device-1 should exist")
	}
	if !store.DeviceExists("device-2") {
		t.Error("device-2 should exist")
	}
	if store.DeviceExists("") {
		t.Error("empty device ID should not exist")
	}
}
