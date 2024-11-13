package integration

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/app"
	"github.com/mochaeng/sapphire-backend/internal/auth"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/mailer"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/ratelimiter"
	service "github.com/mochaeng/sapphire-backend/internal/services"
	redisstore "github.com/mochaeng/sapphire-backend/internal/store/cache/redis"
	"github.com/mochaeng/sapphire-backend/internal/store/postgres"
	"github.com/redis/go-redis/v9"
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
	// logger := zap.NewNop().Sugar()
	logger := zap.Must(zap.NewProduction()).Sugar()
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

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.Auth.Token.Secret,
		cfg.Auth.Token.Issuer,
		cfg.Auth.Token.Issuer,
		cfg.Auth.Token.Expired,
	)

	servicesCfg := &config.ServiceCfg{
		Logger:        logger,
		Store:         store,
		Cfg:           cfg,
		Mailer:        clientMailer,
		Authenticator: jwtAuthenticator,
		CacheStore:    cacheStore,
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
