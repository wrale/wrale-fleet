package service

import (
	"context"
	"fmt"

	"github.com/wrale/wrale-fleet/user/api/types"
)

type deviceService struct {}

// NewDeviceService creates a new device service
func NewDeviceService() types.DeviceService {
	return &deviceService{}
}

func (s *deviceService) List(ctx context.Context) ([]types.Device, error) {
	// TODO: Implement for v1.0
	return nil, fmt.Errorf("not implemented")
}

func (s *deviceService) Get(ctx context.Context, id string) (*types.Device, error) {
	// TODO: Implement for v1.0
	return nil, fmt.Errorf("not implemented")
}

func (s *deviceService) Create(ctx context.Context, device *types.Device) error {
	// TODO: Implement for v1.0
	return fmt.Errorf("not implemented")
}

func (s *deviceService) Update(ctx context.Context, device *types.Device) error {
	// TODO: Implement for v1.0
	return fmt.Errorf("not implemented")
}

func (s *deviceService) Delete(ctx context.Context, id string) error {
	// TODO: Implement for v1.0
	return fmt.Errorf("not implemented")
}

func (s *deviceService) SendCommand(ctx context.Context, id string, cmd *types.DeviceCommand) error {
	// TODO: Implement for v1.0
	return fmt.Errorf("not implemented")
}