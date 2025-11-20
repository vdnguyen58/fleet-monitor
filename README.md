### Prerequisites

- Go 1.24 or higher
- Docker (optional, for containerized deployment)
- Make (optional, for convenience commands)

### TL;DR: How to run

- Docker

```bash
docker run --rm -d --name fleet-monitor -p 6733:6733 ghcr.io/vdnguyen58/fleet-monitor:latest
./device-simulator-linux-amd64
docker stop fleet-monitor
```

- Binary: https://github.com/vdnguyen58/fleet-monitor/releases

```bash
./fleet-monitor-linux-amd64 -csv devices.csv
./device-simulator-linux-amd64
```

- Source code

```bash
git clone https://github.com/vdnguyen58/fleet-monitor.git
go mod download
go run . -csv devices.csv
```

### Running with Go

```bash
# Install dependencies
go mod download

# Run with default settings (port 6733, devices.csv)
go run .

# Run with custom port and CSV file
go run . --csv /path/to/devices.csv

# Or using environment variables
export DEVICES_CSV=/path/to/devices.csv
go run .

# Or using Make
make run
```

### Running with Docker

```bash
# Build and run using Docker directly
docker build -t fleet-monitor .
docker run -p 6733:6733 fleet-monitor

# Or using Make
make docker-build
make docker-run

# View logs
make docker-logs

# Stop container
make docker-stop
```

### Testing

```bash
# Unit-test
make test

# E2E test
make docker-e2e
```

### How long did you spend working on the problem? What was most difficult?

This project was developed through AI-assisted pair programming with Copilot. The implementation took approximately 60 minutes of focused development time, including:

- [Claude code assisted] Boilerplates (15 minutes).
- **Business logic for uptime and upload time calculations (30 minutes)**.
- [Copilot assisted] Comprehensive testing strategy (unit tests + e2e tests) (10 minutes).
- [Copilot assisted] Docker support and CI/CD pipeline (10 minutes)
- [Claude code assisted] Refactoring to table-driven tests (10 minutes).
- README & write-ups.
  - [Claude code assisted] Howtos
  - **Most difficult aspects: edge cases, thread-safe data structure**
  - **Extensibilities**
  - **Scalabilities**
  - **Runtime analysis**

**Most Difficult Aspects:**

1. **Uptime Calculation Algorithm**: The most challenging part was implementing the uptime calculation correctly. The requirement states: *"uptime is calculated as the number of heartbeats received divided by the number of minutes between the first and last heartbeat"*. This required careful handling of edge cases:
   - Single heartbeat should return 100% uptime
   - Time spans less than 1 minute should return 100%
   - Heartbeats can arrive out of order (requires sorting)
   - The calculation can exceed 100% if devices send heartbeats more frequently than once per minute

2. **Thread-Safe Concurrent Access**: Implementing the storage layer with `sync.RWMutex` to safely handle concurrent requests while maintaining performance required careful consideration of lock granularity.

### How would you modify your data model or code to account for more kinds of metrics?

Here's how I would add more metric types:

#### 1. **Extend the Data Model**

```go
// storage/device_store.go
type DeviceData struct {
    DeviceID      string
    Heartbeats    []time.Time
    UploadTimes   []int64

    // New metrics:
    CPUUsage      []CPUMetric
    MemoryUsage   []MemoryMetric
    DiskSpace     []DiskMetric
    NetworkStats  []NetworkMetric
    CustomMetrics map[string][]interface{}  // For arbitrary metrics, such as temperature
}

type CPUMetric struct {
    Timestamp time.Time
    Usage     float64  // percentage
}

type MemoryMetric struct {
    Timestamp time.Time
    Used      int64
    Total     int64
}
```

#### 2. **Implement Metric Registry Pattern**

For better scalability, consider a metric registry:

```go
type MetricCollector interface {
    Collect(deviceID string, data interface{}) error
    Calculate(deviceID string) (interface{}, error)
}

type MetricRegistry struct {
    collectors map[string]MetricCollector
    mu         sync.RWMutex
}

func (r *MetricRegistry) Register(metricType string, collector MetricCollector) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.collectors[metricType] = collector
}
```

#### 3. **Add New API Endpoints**

```go
// Generic metric endpoint
POST /api/v1/devices/{device_id}/metrics/{metric_type}
GET  /api/v1/devices/{device_id}/metrics/{metric_type}

// Specific endpoints
POST /api/v1/devices/{device_id}/cpu
POST /api/v1/devices/{device_id}/memory
GET  /api/v1/devices/{device_id}/metrics/summary
```

#### 4. **Configuration-Driven Metrics**

Use a YAML/JSON config to define metrics:

```yaml
metrics:
  - name: cpu_usage
    type: gauge
    aggregation: avg
    retention: 24h
  - name: error_rate
    type: counter
    aggregation: sum
    retention: 7d
```

### **Time Complexity Analysis**

1. `POST /devices/{device_id}/heartbeat`: O(1)
2. `POST /devices/{device_id}/stats`: O(1)
3. `GET /devices/{device_id}/stats`: O(h log h + u) 
  - h = heartbeats, u = upload stats