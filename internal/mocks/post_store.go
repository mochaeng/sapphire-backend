package mocks

import (
	"context"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockPostStore) GetByUsername(ctontext context.Context, username string, timeCursor time.Time) ([]*models.Post, error) {
	return nil, nil
}
