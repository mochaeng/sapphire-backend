package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/app"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/mailer"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/ratelimiter"
	service "github.com/mochaeng/sapphire-backend/internal/services"
	redisstore "github.com/mochaeng/sapphire-backend/internal/store/cache/redis"
	"github.com/mochaeng/sapphire-backend/internal/store/postgres"
	"github.com/mochaeng/sapphire-backend/internal/testutils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type RegisterUserResponse struct {
	Data models.RegisterUserResponse `json:"data"`
	rr   *httptest.ResponseRecorder
}

var (
	migrationsPath      = fmt.Sprintf("file://%s", filepath.Join("..", "..", "migrate", "migrations"))
	integrationSeedPath = filepath.Join("..", "..", "migrate", "tests", "integration_seed.sql")

	ErrContainerNotStarting     = errors.New("could not create posgres container")
	ErrMigrateDriverNotStarting = errors.New("could not create driver")
	ErrMigrateApplying          = errors.New("could not apply migrations")
	ErrDatabaseSeed             = errors.New("could not seed database")
	ErrPayloadMarshal           = errors.New("could not marshal user struct")
	ErrRequestHTTP              = errors.New("could not make HTTP request")
	ErrResponseParse            = errors.New("could not parse response")
)

func createNewAppSuite(db *sql.DB, parsedRedisConnStr string) (*app.Application, error) {
	logger := zap.NewNop().Sugar()
	// logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	cfg := &config.Cfg{
		Env:         "dev",
		MediaFolder: "data",
		Cacher: config.CacheCfg{
			IsEnable: true,
		},
		Mail: config.MailCfg{
			Expired:   24 * time.Hour,
			FromEmail: "mail@mail.mail",
		},
		Auth: config.AuthCfg{
			Basic: config.BasicAuthCfg{
				Username: "admin",
				Password: "admin",
			},
			Token: config.TokenCfg{
				Secret:  "secrettest",
				Expired: 24 * 7 * time.Hour,
				Issuer:  "sapphiretester",
			},
		},
		RateLimiter: config.RateLimiterConfig{
			RequestPerTimeFrame: 20,
			TimeFrame:           time.Second * 5,
			IsEnable:            true,
		},
	}

	store := postgres.NewPostgresStore(db)

	var rdb *redis.Client
	if cfg.Cacher.IsEnable {
		rdb = redisstore.NewRedisClient(parsedRedisConnStr, "", 0)
	}
	cacheStore := redisstore.NewRedisStore(rdb)

	clientMailer, err := mailer.NewGomail("smtp.google.com", "", "")
	if err != nil {
		return nil, fmt.Errorf("could not create mailer, err: %s", err)
	}

	servicesCfg := &config.ServiceCfg{
		Logger:     logger,
		Store:      store,
		Cfg:        cfg,
		Mailer:     clientMailer,
		CacheStore: cacheStore,
	}
	services := service.NewServices(servicesCfg)

	ratelimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.RateLimiter.RequestPerTimeFrame,
		cfg.RateLimiter.TimeFrame,
	)

	app := &app.Application{
		Config:      cfg,
		Service:     services,
		Logger:      logger,
		RateLimiter: ratelimiter,
	}
	return app, nil
}

func registerUser(t *testing.T, mux http.Handler, payload *models.RegisterUserPayload) RegisterUserResponse {
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, ErrPayloadMarshal)

	req, err := http.NewRequest(http.MethodPost, signupRouter, bytes.NewReader(jsonData))
	require.NoError(t, err, ErrRequestHTTP)

	rr := testutils.ExecuteRequest(req, mux)

	response := RegisterUserResponse{rr: rr}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	return response
}

func makePost(t *testing.T, mux http.Handler, cookie *http.Cookie, tittle, content string, tags []string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("tittle", tittle)
	writer.WriteField("content", content)
	for _, tag := range tags {
		writer.WriteField("tags", tag)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/v1/post/", body)
	require.NoError(t, err, ErrRequestHTTP)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)

	rr := testutils.ExecuteRequest(req, mux)
	var resp struct {
		Data models.CreatePostResponse `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func activateUser(t *testing.T, mux http.Handler, token string) {
	activationURL := fmt.Sprintf("%s%s", activateRouter, token)
	req, err := http.NewRequest(http.MethodPut, activationURL, nil)
	require.NoError(t, err, ErrRequestHTTP)
	rr := testutils.ExecuteRequest(req, mux)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func signinUser(t *testing.T, mux http.Handler, email string, password string) *httptest.ResponseRecorder {
	payload := models.SigninPayload{
		Email:    email,
		Password: password,
	}
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, ErrPayloadMarshal)
	req, err := http.NewRequest(http.MethodPost, signinRouter, bytes.NewReader(jsonData))
	require.NoError(t, err, ErrRequestHTTP)
	rr := testutils.ExecuteRequest(req, mux)
	assert.Equal(t, http.StatusCreated, rr.Code)
	return rr
}
