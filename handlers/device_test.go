package handlers

import (
	"math"
	"testing"
	"time"
)

// ---------------------------------------
// Uptime test
// ---------------------------------------

func TestCalculateUptime(t *testing.T) {
	baseTime := time.Now()

	testCases := []struct {
		name       string
		heartbeats []time.Time
		expected   float64
		tolerance  float64 // For floating point comparisons
	}{
		{
			name:       "No heartbeats",
			heartbeats: []time.Time{},
			expected:   0.0,
			tolerance:  0.0,
		},
		{
			name:       "Single heartbeat",
			heartbeats: []time.Time{baseTime},
			expected:   100.0,
			tolerance:  0.0,
		},
		{
			name: "Perfect uptime - 5 heartbeats over 4 minutes",
			heartbeats: []time.Time{
				baseTime,
				baseTime.Add(1 * time.Minute),
				baseTime.Add(2 * time.Minute),
				baseTime.Add(3 * time.Minute),
				baseTime.Add(4 * time.Minute),
			},
			expected:  (5.0 / 4.0) * 100.0,
			tolerance: 0.0,
		},
		{
			name: "Missed heartbeats - 3 over 5 minutes",
			heartbeats: []time.Time{
				baseTime,
				baseTime.Add(1 * time.Minute),
				baseTime.Add(5 * time.Minute),
			},
			expected:  (3.0 / 5.0) * 100.0,
			tolerance: 0.0,
		},
		{
			name: "Unsorted heartbeats",
			heartbeats: []time.Time{
				baseTime.Add(3 * time.Minute),
				baseTime,
				baseTime.Add(1 * time.Minute),
				baseTime.Add(2 * time.Minute),
			},
			expected:  (4.0 / 3.0) * 100.0,
			tolerance: 0.000001,
		},
		{
			name: "Less than one minute duration",
			heartbeats: []time.Time{
				baseTime,
				baseTime.Add(30 * time.Second),
			},
			expected:  100.0,
			tolerance: 0.0,
		},
		{
			name: "Exactly one minute",
			heartbeats: []time.Time{
				baseTime,
				baseTime.Add(1 * time.Minute),
			},
			expected:  (2.0 / 1.0) * 100.0,
			tolerance: 0.0,
		},
		{
			name: "Ten minutes perfect",
			heartbeats: func() []time.Time {
				hbs := make([]time.Time, 10)
				for i := 0; i < 10; i++ {
					hbs[i] = baseTime.Add(time.Duration(i) * time.Minute)
				}
				return hbs
			}(),
			expected:  (10.0 / 9.0) * 100.0,
			tolerance: 0.000001,
		},
		{
			name: "Half missed - heartbeats every 2 minutes",
			heartbeats: []time.Time{
				baseTime,
				baseTime.Add(2 * time.Minute),
				baseTime.Add(4 * time.Minute),
				baseTime.Add(6 * time.Minute),
				baseTime.Add(8 * time.Minute),
				baseTime.Add(10 * time.Minute),
			},
			expected:  (6.0 / 10.0) * 100.0,
			tolerance: 0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uptime := calculateUptime(tc.heartbeats)
			if tc.tolerance > 0 {
				if math.Abs(uptime-tc.expected) > tc.tolerance {
					t.Errorf("Expected %f, got %f", tc.expected, uptime)
				}
			} else {
				if uptime != tc.expected {
					t.Errorf("Expected %f, got %f", tc.expected, uptime)
				}
			}
		})
	}
}

// ---------------------------------------
// Average upload time test
// ---------------------------------------

func TestCalculateAvgUploadTime(t *testing.T) {
	testCases := []struct {
		name        string
		uploadTimes []int64
		expected    string
	}{
		{
			name:        "No data",
			uploadTimes: []int64{},
			expected:    "0s",
		},
		{
			name:        "Single value - 5 seconds",
			uploadTimes: []int64{5 * time.Second.Nanoseconds()},
			expected:    "5s",
		},
		{
			name: "Multiple values - avg of 5s and 10s",
			uploadTimes: []int64{
				5 * time.Second.Nanoseconds(),
				10 * time.Second.Nanoseconds(),
			},
			expected: "7.5s",
		},
		{
			name: "Three values - 3s, 6s, 9s",
			uploadTimes: []int64{
				3 * time.Second.Nanoseconds(),
				6 * time.Second.Nanoseconds(),
				9 * time.Second.Nanoseconds(),
			},
			expected: "6s",
		},
		{
			name: "Minutes - 1min, 2min, 3min",
			uploadTimes: []int64{
				60 * time.Second.Nanoseconds(),
				120 * time.Second.Nanoseconds(),
				180 * time.Second.Nanoseconds(),
			},
			expected: "2m0s",
		},
		{
			name: "Milliseconds - 1ms, 2ms, 3ms",
			uploadTimes: []int64{
				1 * time.Millisecond.Nanoseconds(),
				2 * time.Millisecond.Nanoseconds(),
				3 * time.Millisecond.Nanoseconds(),
			},
			expected: "2ms",
		},
		{
			name: "Mixed milliseconds - 100ms, 200ms, 300ms",
			uploadTimes: []int64{
				100 * time.Millisecond.Nanoseconds(),
				200 * time.Millisecond.Nanoseconds(),
				300 * time.Millisecond.Nanoseconds(),
			},
			expected: "200ms",
		},
		{
			name: "Hours and minutes - 30min, 1hr, 1.5hr",
			uploadTimes: []int64{
				30 * time.Minute.Nanoseconds(),
				60 * time.Minute.Nanoseconds(),
				90 * time.Minute.Nanoseconds(),
			},
			expected: "1h0m0s",
		},
		{
			name: "Very small - microseconds",
			uploadTimes: []int64{
				1000, // 1 microsecond
				2000, // 2 microseconds
				3000, // 3 microseconds
			},
			expected: "2µs",
		},
		{
			name: "Odd average - 2s and 3s",
			uploadTimes: []int64{
				2 * time.Second.Nanoseconds(),
				3 * time.Second.Nanoseconds(),
			},
			expected: "2.5s",
		},
		{
			name: "Large dataset - 10 values of 5s",
			uploadTimes: func() []int64 {
				times := make([]int64, 10)
				for i := 0; i < 10; i++ {
					times[i] = 5 * time.Second.Nanoseconds()
				}
				return times
			}(),
			expected: "5s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			avgTime := calculateAvgUploadTime(tc.uploadTimes)
			if avgTime != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, avgTime)
			}
		})
	}
}
