package coordinator

import (
	"context"
	"net/http"

	"github.com/wrale/wrale-fleet/fleet/types"
)

// defaultMetalClient implements the MetalClient interface
type defaultMetalClient struct {
	client  *http.Client
	baseURL string
}

// NewMetalClient creates a new metal service client
func NewMetalClient(baseURL string) MetalClient {
	return &defaultMetalClient{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (m *defaultMetalClient) ExecuteOperation(ctx context.Context, deviceID types.DeviceID, operation string) error {
	// TODO: Implement metal API call
	return nil
}

func (m *defaultMetalClient) GetDeviceMetrics(ctx context.Context, deviceID types.DeviceID) (*types.DeviceMetrics, error) {
	// TODO: Implement metal API call
	return &types.DeviceMetrics{}, nil
}
