package server

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// Mock services for testing

type mockDeviceService struct {
    devices map[types.DeviceID]*apitypes.DeviceResponse
}

func newMockDeviceService() *mockDeviceService {
    return &mockDeviceService{
        devices: make(map[types.DeviceID]*apitypes.DeviceResponse),
    }
}

func (m *mockDeviceService) CreateDevice(req *apitypes.DeviceCreateRequest) (*apitypes.DeviceResponse, error) {
    resp := &apitypes.DeviceResponse{
        ID:         req.ID,
        Status:     "active",
        Location:   req.Location,
        Config:     req.Config,
        LastUpdate: time.Now(),
    }
    m.devices[req.ID] = resp
    return resp, nil
}

func (m *mockDeviceService) GetDevice(id types.DeviceID) (*apitypes.DeviceResponse, error) {
    if device, exists := m.devices[id]; exists {
        return device, nil
    }
    return nil, fmt.Errorf("device not found")
}

func (m *mockDeviceService) UpdateDevice(id types.DeviceID, req *apitypes.DeviceUpdateRequest) (*apitypes.DeviceResponse, error) {
    if device, exists := m.devices[id]; exists {
        if req.Status != "" {
            device.Status = req.Status
        }
        if req.Location != nil {
            device.Location = *req.Location
        }
        if req.Config != nil {
            device.Config = req.Config
        }
        device.LastUpdate = time.Now()
        return device, nil
    }
    return nil, fmt.Errorf("device not found")
}

func (m *mockDeviceService) ListDevices() ([]*apitypes.DeviceResponse, error) {
    devices := make([]*apitypes.DeviceResponse, 0)
    for _, device := range m.devices {
        devices = append(devices, device)
    }
    return devices, nil
}

func (m *mockDeviceService) DeleteDevice(id types.DeviceID) error {
    delete(m.devices, id)
    return nil
}

func (m *mockDeviceService) ExecuteCommand(id types.DeviceID, req *apitypes.DeviceCommandRequest) (*apitypes.CommandResponse, error) {
    return &apitypes.CommandResponse{
        ID:        "test-command",
        Status:    "completed",
        StartTime: time.Now(),
        EndTime:   nil,
    }, nil
}

type mockFleetService struct{}

func (m *mockFleetService) ExecuteFleetCommand(req *apitypes.FleetCommandRequest) (*apitypes.CommandResponse, error) {
    return &apitypes.CommandResponse{
        ID:        "test-fleet-command",
        Status:    "completed",
        StartTime: time.Now(),
    }, nil
}

func (m *mockFleetService) GetFleetMetrics() (map[string]interface{}, error) {
    return map[string]interface{}{
        "total_devices": 1,
        "active_devices": 1,
    }, nil
}

func (m *mockFleetService) UpdateConfig(req *apitypes.ConfigUpdateRequest) error {
    return nil
}

func (m *mockFleetService) GetConfig(devices []types.DeviceID) (map[types.DeviceID]map[string]interface{}, error) {
    return map[types.DeviceID]map[string]interface{}{
        "test-device": {
            "setting": "value",
        },
    }, nil
}

type mockAuthService struct{}

func (m *mockAuthService) Authenticate(token string) (bool, error) {
    return token == "valid-token", nil
}

func (m *mockAuthService) Authorize(token, resource, action string) (bool, error) {
    return token == "valid-token", nil
}

func (m *mockAuthService) GenerateToken(userID string, roles []string) (string, error) {
    return "valid-token", nil
}

func TestServer(t *testing.T) {
    // Create test server with mock services
    deviceSvc := newMockDeviceService()
    fleetSvc := &mockFleetService{}
    authSvc := &mockAuthService{}
    server := NewServer(deviceSvc, fleetSvc, nil, authSvc)

    t.Run("Authentication", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/devices", nil)
        resp := httptest.NewRecorder()

        // No token
        server.router.ServeHTTP(resp, req)
        if resp.Code != http.StatusUnauthorized {
            t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
        }

        // Invalid token
        req.Header.Set("Authorization", "invalid-token")
        resp = httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)
        if resp.Code != http.StatusUnauthorized {
            t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
        }
    })

    t.Run("Device CRUD", func(t *testing.T) {
        // Create device
        createReq := apitypes.DeviceCreateRequest{
            ID: "test-device",
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 1,
                Zone:     "zone-1",
            },
            Config: map[string]interface{}{
                "setting": "value",
            },
        }

        body, _ := json.Marshal(createReq)
        req := httptest.NewRequest("POST", "/api/v1/devices", bytes.NewReader(body))
        req.Header.Set("Authorization", "valid-token")
        resp := httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusCreated {
            t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.Code)
        }

        // Get device
        req = httptest.NewRequest("GET", "/api/v1/devices/test-device", nil)
        req.Header.Set("Authorization", "valid-token")
        resp = httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusOK {
            t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
        }

        // Update device
        updateReq := apitypes.DeviceUpdateRequest{
            Status: "standby",
        }
        body, _ = json.Marshal(updateReq)
        req = httptest.NewRequest("PUT", "/api/v1/devices/test-device", bytes.NewReader(body))
        req.Header.Set("Authorization", "valid-token")
        resp = httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusOK {
            t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
        }

        // Delete device
        req = httptest.NewRequest("DELETE", "/api/v1/devices/test-device", nil)
        req.Header.Set("Authorization", "valid-token")
        resp = httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusNoContent {
            t.Errorf("Expected status %d, got %d", http.StatusNoContent, resp.Code)
        }
    })

    t.Run("Device Commands", func(t *testing.T) {
        // Create test device
        deviceID := types.DeviceID("command-test-device")
        deviceSvc.devices[deviceID] = &apitypes.DeviceResponse{
            ID:     deviceID,
            Status: "active",
        }

        // Execute command
        commandReq := apitypes.DeviceCommandRequest{
            Operation: "test_operation",
            Params: map[string]interface{}{
                "param": "value",
            },
        }

        body, _ := json.Marshal(commandReq)
        req := httptest.NewRequest("POST", "/api/v1/devices/command-test-device/command", bytes.NewReader(body))
        req.Header.Set("Authorization", "valid-token")
        resp := httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusOK {
            t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
        }
    })

    t.Run("Fleet Operations", func(t *testing.T) {
        // Fleet command
        fleetReq := apitypes.FleetCommandRequest{
            Operation: "test_operation",
            Devices:   []types.DeviceID{"device-1", "device-2"},
        }

        body, _ := json.Marshal(fleetReq)
        req := httptest.NewRequest("POST", "/api/v1/fleet/command", bytes.NewReader(body))
        req.Header.Set("Authorization", "valid-token")
        resp := httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusOK {
            t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
        }

        // Fleet metrics
        req = httptest.NewRequest("GET", "/api/v1/fleet/metrics", nil)
        req.Header.Set("Authorization", "valid-token")
        resp = httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusOK {
            t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
        }
    })

    t.Run("Configuration", func(t *testing.T) {
        // Update config
        configReq := apitypes.ConfigUpdateRequest{
            Config: map[string]interface{}{
                "setting": "value",
            },
            Devices: []types.DeviceID{"device-1"},
        }

        body, _ := json.Marshal(configReq)
        req := httptest.NewRequest("PUT", "/api/v1/fleet/config", bytes.NewReader(body))
        req.Header.Set("Authorization", "valid-token")
        resp := httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusNoContent {
            t.Errorf("Expected status %d, got %d", http.StatusNoContent, resp.Code)
        }

        // Get config
        req = httptest.NewRequest("GET", "/api/v1/fleet/config?device=test-device", nil)
        req.Header.Set("Authorization", "valid-token")
        resp = httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusOK {
            t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
        }
    })

    t.Run("Health Check", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/health", nil)
        req.Header.Set("Authorization", "valid-token")
        resp := httptest.NewRecorder()
        server.router.ServeHTTP(resp, req)

        if resp.Code != http.StatusOK {
            t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
        }

        var response apitypes.APIResponse
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            t.Errorf("Failed to decode response: %v", err)
        }

        data := response.Data.(map[string]interface{})
        if status := data["status"]; status != "healthy" {
            t.Errorf("Expected status 'healthy', got '%v'", status)
        }
    })
}
