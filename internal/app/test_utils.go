package app

import (
	"net/http"
	"net/http/httptest"
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

	mockStore := mocks.NewMockStore()
	mockService := mocks.NewMockService(&mockStore)

	return &Application{
		Config:  &config.Cfg{},
		Service: &mockService,
		Logger:  logger,
	}
}

func executeRequest(r *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, r)
	return rr
}
