package mocks

import (
	service "github.com/mochaeng/sapphire-backend/internal/services"
)

func NewMockService() service.Service {
	return service.Service{
		User: &MockUserService{},
		Post: &MockPostService{},
	}
}
