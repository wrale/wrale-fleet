package service

import (
	"context"
	"fmt"

	"github.com/wrale/wrale-fleet/user/api/types"
)

type fleetService struct {}

// NewFleetService creates a new fleet service
func NewFleetService() types.FleetService {
	return &fleetService{}
}

func (s *fleetService) GetMetrics(ctx context.Context) (*types.FleetMetrics, error) {
	// TODO: Implement for v1.0
	return nil, fmt.Errorf("not implemented")
}

func (s *fleetService) GetConfig(ctx context.Context) (*types.FleetConfig, error) {
	// TODO: Implement for v1.0
	return nil, fmt.Errorf("not implemented")
}

func (s *fleetService) UpdateConfig(ctx context.Context, config *types.FleetConfig) error {
	// TODO: Implement for v1.0
	return fmt.Errorf("not implemented")
}

func (s *fleetService) SendCommand(ctx context.Context, cmd *types.FleetCommand) error {
	// TODO: Implement for v1.0
	return fmt.Errorf("not implemented")
}