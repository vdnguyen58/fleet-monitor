package handlers

import (
	"math"
	"testing"
	"time"
)


// ---------------------------------------
// Uptime test
// ---------------------------------------

func TestCalculateUptime_NoHeartbeats(t *testing.T) {
	uptime := calculateUptime([]time.Time{})
	if uptime != 0.0 {
		t.Errorf("Expected 0.0, got %f", uptime)
	}
}

func TestCalculateUptime_SingleHeartbeat(t *testing.T) {
	heartbeats := []time.Time{time.Now()}
	uptime := calculateUptime(heartbeats)
	if uptime != 100.0 {
		t.Errorf("Expected 100.0 for single heartbeat, got %f", uptime)
	}
}

func TestCalculateUptime_PerfectUptime(t *testing.T) {
	// Device sends heartbeat every minute for 5 minutes
	baseTime := time.Now()
	heartbeats := []time.Time{
		baseTime,
		baseTime.Add(1 * time.Minute),
		baseTime.Add(2 * time.Minute),
		baseTime.Add(3 * time.Minute),
		baseTime.Add(4 * time.Minute),
	}

	uptime := calculateUptime(heartbeats)

	// 5 heartbeats over 4 minutes = (5/4) * 100 = 125%
	expected := (5.0 / 4.0) * 100.0
	if uptime != expected {
		t.Errorf("Expected %f, got %f", expected, uptime)
	}
}

func TestCalculateUptime_MissedHeartbeats(t *testing.T) {
	// Device misses some heartbeats: 3 heartbeats over 5 minutes
	baseTime := time.Now()
	heartbeats := []time.Time{
		baseTime,
		baseTime.Add(1 * time.Minute),
		baseTime.Add(5 * time.Minute), // Missed 2, 3, 4
	}

	uptime := calculateUptime(heartbeats)

	// 3 heartbeats over 5 minutes = (3/5) * 100 = 60%
	expected := (3.0 / 5.0) * 100.0
	if uptime != expected {
		t.Errorf("Expected %f, got %f", expected, uptime)
	}
}

func TestCalculateUptime_UnsortedHeartbeats(t *testing.T) {
	// Heartbeats provided in random order
	baseTime := time.Now()
	heartbeats := []time.Time{
		baseTime.Add(3 * time.Minute),
		baseTime,
		baseTime.Add(1 * time.Minute),
		baseTime.Add(2 * time.Minute),
	}

	uptime := calculateUptime(heartbeats)

	// Should correctly sort and calculate: 4 heartbeats over 3 minutes
	expected := (4.0 / 3.0) * 100.0
	if math.Abs(uptime-expected) > 0.000001 {
		t.Errorf("Expected %f, got %f", expected, uptime)
	}
}

func TestCalculateUptime_LessThanOneMinute(t *testing.T) {
	baseTime := time.Now()
	heartbeats := []time.Time{
		baseTime,
		baseTime.Add(30 * time.Second), // Less than 1 minute apart
	}

	uptime := calculateUptime(heartbeats)

	// Less than 1 minute duration should return 100% uptime
	if uptime != 100.0 {
		t.Errorf("Expected 100.0 for duration < 1 minute, got %f", uptime)
	}
}

func TestCalculateUptime_ExactlyOneMinute(t *testing.T) {
	baseTime := time.Now()
	heartbeats := []time.Time{
		baseTime,
		baseTime.Add(1 * time.Minute),
	}

	uptime := calculateUptime(heartbeats)

	// 2 heartbeats over 1 minute = (2/1) * 100 = 200%
	expected := (2.0 / 1.0) * 100.0
	if uptime != expected {
		t.Errorf("Expected %f, got %f", expected, uptime)
	}
}

func TestCalculateUptime_TenMinutesPerfect(t *testing.T) {
	// Heartbeat every minute for 10 minutes
	baseTime := time.Now()
	heartbeats := make([]time.Time, 10)
	for i := 0; i < 10; i++ {
		heartbeats[i] = baseTime.Add(time.Duration(i) * time.Minute)
	}

	uptime := calculateUptime(heartbeats)

	// 10 heartbeats over 9 minutes = (10/9) * 100 = 111.111...%
	expected := (10.0 / 9.0) * 100.0
	if math.Abs(uptime-expected) > 0.000001 {
		t.Errorf("Expected %f, got %f", expected, uptime)
	}
}

func TestCalculateUptime_HalfMissed(t *testing.T) {
	// Device sends heartbeat every 2 minutes for 10 minutes
	// This means half the heartbeats are missing
	baseTime := time.Now()
	heartbeats := []time.Time{
		baseTime,
		baseTime.Add(2 * time.Minute),
		baseTime.Add(4 * time.Minute),
		baseTime.Add(6 * time.Minute),
		baseTime.Add(8 * time.Minute),
		baseTime.Add(10 * time.Minute),
	}

	uptime := calculateUptime(heartbeats)

	// 6 heartbeats over 10 minutes = (6/10) * 100 = 60%
	expected := (6.0 / 10.0) * 100.0
	if uptime != expected {
		t.Errorf("Expected %f, got %f", expected, uptime)
	}
}

// ---------------------------------------
// Average upload time test
// ---------------------------------------

func TestCalculateAvgUploadTime_NoData(t *testing.T) {
	avgTime := calculateAvgUploadTime([]int64{})
	if avgTime != "0s" {
		t.Errorf("Expected '0s' for no data, got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_SingleValue(t *testing.T) {
	// Single upload time of 5 seconds
	uploadTimes := []int64{5 * time.Second.Nanoseconds()}
	avgTime := calculateAvgUploadTime(uploadTimes)

	if avgTime != "5s" {
		t.Errorf("Expected '5s', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_MultipleValues(t *testing.T) {
	// Average of 5s and 10s = 7.5s
	uploadTimes := []int64{
		5 * time.Second.Nanoseconds(),
		10 * time.Second.Nanoseconds(),
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	if avgTime != "7.5s" {
		t.Errorf("Expected '7.5s', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_ThreeValues(t *testing.T) {
	// Average of 3s, 6s, 9s = 6s
	uploadTimes := []int64{
		3 * time.Second.Nanoseconds(),
		6 * time.Second.Nanoseconds(),
		9 * time.Second.Nanoseconds(),
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	if avgTime != "6s" {
		t.Errorf("Expected '6s', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_Minutes(t *testing.T) {
	// Average of 1min, 2min, 3min = 2min
	uploadTimes := []int64{
		60 * time.Second.Nanoseconds(),
		120 * time.Second.Nanoseconds(),
		180 * time.Second.Nanoseconds(),
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	if avgTime != "2m0s" {
		t.Errorf("Expected '2m0s', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_Milliseconds(t *testing.T) {
	// Average of 1ms, 2ms, 3ms = 2ms
	uploadTimes := []int64{
		1 * time.Millisecond.Nanoseconds(),
		2 * time.Millisecond.Nanoseconds(),
		3 * time.Millisecond.Nanoseconds(),
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	if avgTime != "2ms" {
		t.Errorf("Expected '2ms', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_Mixed(t *testing.T) {
	// Mix of small and large values
	uploadTimes := []int64{
		100 * time.Millisecond.Nanoseconds(), // 100ms
		200 * time.Millisecond.Nanoseconds(), // 200ms
		300 * time.Millisecond.Nanoseconds(), // 300ms
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	// Average = 200ms
	if avgTime != "200ms" {
		t.Errorf("Expected '200ms', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_HoursAndMinutes(t *testing.T) {
	// Average of 30min, 1hr, 1.5hr = 1hr
	uploadTimes := []int64{
		30 * time.Minute.Nanoseconds(),
		60 * time.Minute.Nanoseconds(),
		90 * time.Minute.Nanoseconds(),
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	if avgTime != "1h0m0s" {
		t.Errorf("Expected '1h0m0s', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_VerySmall(t *testing.T) {
	// Microseconds
	uploadTimes := []int64{
		1000,  // 1 microsecond
		2000,  // 2 microseconds
		3000,  // 3 microseconds
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	// Average = 2 microseconds
	if avgTime != "2µs" {
		t.Errorf("Expected '2µs', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_OddAverage(t *testing.T) {
	// Values that result in odd average
	uploadTimes := []int64{
		2 * time.Second.Nanoseconds(),
		3 * time.Second.Nanoseconds(),
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	// Average = 2.5s
	if avgTime != "2.5s" {
		t.Errorf("Expected '2.5s', got '%s'", avgTime)
	}
}

func TestCalculateAvgUploadTime_LargeDataset(t *testing.T) {
	// 10 values all 5 seconds
	uploadTimes := make([]int64, 10)
	for i := 0; i < 10; i++ {
		uploadTimes[i] = 5 * time.Second.Nanoseconds()
	}
	avgTime := calculateAvgUploadTime(uploadTimes)

	if avgTime != "5s" {
		t.Errorf("Expected '5s', got '%s'", avgTime)
	}
}
