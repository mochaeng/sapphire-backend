package mocks

import (
	"github.com/mochaeng/sapphire-backend/internal/store"
)

func NewMockStore() store.Store {
	return store.Store{
		Post: &MockPostStore{},
	}
}
