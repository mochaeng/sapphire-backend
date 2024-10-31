package mocks

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockStore() store.Store {
	return store.Store{
		Post: &MockPostStore{},
	}
}

type MockPostStore struct {
	mock.Mock
}

func (m *MockPostStore) Create(context.Context, *models.Post) error {
	return nil
}

func (m *MockPostStore) GetByID(context.Context, int64) (*models.Post, error) {
	return nil, nil
}

func (m *MockPostStore) GetByIDWithUser(context.Context, int64) (*models.Post, error) {
	return nil, nil
}

func (m *MockPostStore) DeleteByID(context.Context, int64) error {
	return nil
}

func (m *MockPostStore) UpdateByID(context.Context, *models.Post) error {
	return nil
}

func (m *MockPostStore) GetAllByUsername(context.Context, string) ([]models.Post, error) {
	return nil, nil
}
