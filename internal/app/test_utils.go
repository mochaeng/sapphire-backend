package app

import (
	"testing"

	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/mocks"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *Application {
	t.Helper()

	logger := zap.NewNop().Sugar()
	// logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	mockService := mocks.NewMockService()

	return &Application{
		Config:  &config.Cfg{},
		Service: &mockService,
		Logger:  logger,
	}
}
