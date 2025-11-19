package storage

import (
	"os"
	"testing"
)

func TestLoadDevicesFromCSV(t *testing.T) {
	testCases := []struct {
		name            string
		csvContent      string
		useTempFile     bool
		filepath        string
		expectError     bool
		expectedDevices []string
		unexpectedIDs   []string
	}{
		{
			name: "Success - load three devices",
			csvContent: `device_id
60-6b-44-84-dc-64
b4-45-52-a2-f1-3c
26-9a-66-01-33-83
`,
			useTempFile: true,
			expectError: false,
			expectedDevices: []string{
				"60-6b-44-84-dc-64",
				"b4-45-52-a2-f1-3c",
				"26-9a-66-01-33-83",
			},
		},
		{
			name:        "File not found",
			useTempFile: false,
			filepath:    "nonexistent-file.csv",
			expectError: true,
		},
		{
			name: "Empty file - only header",
			csvContent: `device_id
`,
			useTempFile:     true,
			expectError:     false,
			expectedDevices: []string{},
			unexpectedIDs:   []string{""},
		},
		{
			name: "With empty lines",
			csvContent: `device_id
device-1

device-2

`,
			useTempFile: true,
			expectError: false,
			expectedDevices: []string{
				"device-1",
				"device-2",
			},
			unexpectedIDs: []string{""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filepath string

			// Setup: Create temp file if needed
			if tc.useTempFile {
				tmpFile, err := os.CreateTemp("", "devices-*.csv")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(tmpFile.Name())

				if _, err := tmpFile.WriteString(tc.csvContent); err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}
				tmpFile.Close()
				filepath = tmpFile.Name()
			} else {
				filepath = tc.filepath
			}

			// Execute
			store := NewDeviceStore()
			err := store.LoadDevicesFromCSV(filepath)

			// Validate error expectation
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadDevicesFromCSV failed: %v", err)
			}

			// Validate expected devices exist
			for _, deviceID := range tc.expectedDevices {
				if !store.DeviceExists(deviceID) {
					t.Errorf("Expected device %s to exist", deviceID)
				}
			}

			// Validate unexpected device IDs don't exist
			for _, deviceID := range tc.unexpectedIDs {
				if store.DeviceExists(deviceID) {
					t.Errorf("Did not expect device '%s' to exist", deviceID)
				}
			}
		})
	}
}
