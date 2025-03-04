package mocks

import "github.com/mochaeng/sapphire-backend/internal/services"

func NewMockService() services.Service {
	return services.Service{
		User: &MockUserService{},
		Post: &MockPostService{},
	}
}
