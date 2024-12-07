package service

import (
	"github.com/wrale/wrale-fleet/user/api/types"
)

type authService struct {}

// NewAuthService creates a new authentication service
func NewAuthService() types.AuthService {
	return &authService{}
}

func (s *authService) Authenticate(token string) (bool, error) {
	// TODO: Implement real authentication for v1.0
	return true, nil
}

func (s *authService) Authorize(token string, path string, method string) (bool, error) {
	// TODO: Implement real authorization for v1.0
	return true, nil
}