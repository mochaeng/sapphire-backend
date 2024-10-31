package main

import (
	"expvar"
	"os"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/mochaeng/sapphire-backend/internal/app"
	"github.com/mochaeng/sapphire-backend/internal/auth"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/database"
	"github.com/mochaeng/sapphire-backend/internal/env"
	"github.com/mochaeng/sapphire-backend/internal/mailer"
	"github.com/mochaeng/sapphire-backend/internal/ratelimiter"
	service "github.com/mochaeng/sapphire-backend/internal/services"
	redisstore "github.com/mochaeng/sapphire-backend/internal/store/cache/redis"
	"github.com/mochaeng/sapphire-backend/internal/store/postgres"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

//	@title			Sapphire API
//	@description	API for Sapphire
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securitydefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	err := godotenv.Load()
	if err != nil {
		logger.Fatalw("error loading .env file", "err", err)
	}

	cfg := &config.Cfg{
		Addr: env.GetString("ADDR", ":7777"),
		DbConfig: config.DbCfg{
			Addr: env.GetString(
				"DATABASE_ADDR",
				"postgres://hutao:adminpassword@localhost:8888/limerence?sslmode=disable",
			),
			MaxOpenConns:       env.GetInt("DATABASE_MAX_OPEN_CONNS", 30),
			MaxIdleConns:       env.GetInt("DATABASE_MAX_IDLE_CONNS", 30),
			MaxConnIdleSeconds: env.GetInt("DATABASE_MAX_CONN_IDLE_SECONDS", 900),
		},
		Cacher: config.CacheCfg{
			IsEnable: env.GetBool("CACHER_IS_ENABLE", true),
			Redis: config.RedisCfg{
				Addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
				Password: env.GetString("REDIS_PASSWORD", ""),
				Db:       env.GetInt("REDIS_DB", 0),
			},
		},
		Env:         "dev",
		Version:     "0.0.1",
		MediaFolder: "data",
		ApiURL:      env.GetString("EXTERNAL_URL", "localhost:7777"),
		FrontedURL:  env.GetString("FRONTED_URL", "http://localhost:3000"),
		Mail: config.MailCfg{
			Expired:   24 * time.Hour,
			FromEmail: env.GetString("FROM_EMAIL", ""),
		},
		Auth: config.AuthCfg{
			Basic: config.BasicAuthCfg{
				Username: env.GetString("AUTH_BASIC_USER", "admin"),
				Password: env.GetString("AUTH_BASIC_PASSWORD", "admin"),
			},
			Token: config.TokenCfg{
				Secret:  env.GetString("JWT_SECRET", ""),
				Expired: 24 * 7 * time.Hour,
				Issuer:  "sapphire",
			},
		},
		RateLimiter: config.RateLimiterConfig{
			RequestPerTimeFrame: env.GetInt("RATE_LIMITER_REQUESTS_COUNT", 20),
			TimeFrame:           time.Second * 5,
			IsEnable:            env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	if _, err := os.Stat(cfg.MediaFolder); os.IsNotExist(err) {
		err := os.MkdirAll(cfg.MediaFolder, os.ModePerm)
		if err != nil {
			logger.Fatalw("not possible to create media folder directory", "err", err)
		}
		logger.Info("media folder was created", "path", cfg.MediaFolder)
	}

	db, err := database.New(
		cfg.DbConfig.Addr,
		cfg.DbConfig.MaxOpenConns,
		cfg.DbConfig.MaxIdleConns,
		cfg.DbConfig.MaxConnIdleSeconds,
	)
	if err != nil {
		logger.Panicw("could not start database connection", "error", err)
	}
	defer db.Close()
	logger.Info("database connection pool established")
	store := postgres.NewPostgresStore(db)

	var rdb *redis.Client
	if cfg.Cacher.IsEnable {
		rdb = redisstore.NewRedisClient(
			cfg.Cacher.Redis.Addr,
			cfg.Cacher.Redis.Password,
			cfg.Cacher.Redis.Db,
		)
		logger.Info("redis cache connection established")
	}
	cacheStore := redisstore.NewRedisStore(rdb)

	smtpServer := env.GetString("SMTP_SERVER", "")
	fromEmail := env.GetString("FROM_EMAIL", "")
	emailPassword := env.GetString("EMAIL_PASSWORD", "")
	clientMailer, err := mailer.NewGomail(smtpServer, fromEmail, emailPassword)
	if err != nil {
		logger.Panicw("could not create mailer", "error", err)
	}

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.Auth.Token.Secret,
		cfg.Auth.Token.Issuer,
		cfg.Auth.Token.Issuer,
		cfg.Auth.Token.Expired,
	)

	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.RateLimiter.RequestPerTimeFrame,
		cfg.RateLimiter.TimeFrame,
	)

	serviceCfg := config.ServiceCfg{
		Logger:        logger,
		Store:         store,
		Cfg:           cfg,
		Mailer:        clientMailer,
		Authenticator: jwtAuthenticator,
		CacheStore:    cacheStore,
	}
	services := service.NewServices(&serviceCfg)

	app := &app.Application{
		Config:      cfg,
		Service:     services,
		Logger:      logger,
		RateLimiter: rateLimiter,
	}

	expvar.NewString("version").Set(cfg.Version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.Mount()
	app.Logger.Fatal(app.Run(mux))
}
