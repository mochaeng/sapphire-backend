package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/mochaeng/sapphire-backend/internal/app"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/cronjobs"
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

var migrationsPath = fmt.Sprintf("file://%s", filepath.Join("migrate", "migrations"))

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

	godotenv.Load()

	cfg := &config.Cfg{
		Addr:    env.GetString("ADDR", "localhost:7777"),
		AppName: "Sapphire",
		DbConfig: config.DbCfg{
			Addr: env.GetString(
				"DATABASE_URL",
				"postgres://hutao:adminpassword@localhost:8888/limerence?sslmode=disable",
			),
			MaxOpenConns:       env.GetInt("DATABASE_MAX_OPEN_CONNS", 30),
			MaxIdleConns:       env.GetInt("DATABASE_MAX_IDLE_CONNS", 30),
			MaxConnIdleSeconds: env.GetInt("DATABASE_MAX_CONN_IDLE_SECONDS", 900),
		},
		Cacher: config.CacheCfg{
			IsEnable: env.GetBool("CACHER_IS_ENABLE", true),
			Redis: config.RedisCfg{
				Addr:     env.GetString("REDIS_ADDR", "redis:6379"),
				Password: env.GetString("REDIS_PASSWORD", ""),
				Db:       env.GetInt("REDIS_DB", 0),
			},
		},
		Env:         env.GetString("ENV", "dev"),
		Version:     "0.0.1",
		MediaFolder: "data",
		FrontedURL:  env.GetString("FRONTED_URL", "http://localhost:5173"),
		Mail: config.MailCfg{
			Expired:   1 * time.Minute,
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
		OAuth: config.OAuthConfig{
			Google: config.GoogleOAuth{
				Key:         env.GetString("GOOGLE_KEY", ""),
				Secret:      env.GetString("GOOGLE_SECRET", ""),
				CallbackURI: env.GetString("GOOGLE_CALLBACK_URI", ""),
			},
		},
	}

	// media folder
	if _, err := os.Stat(cfg.MediaFolder); os.IsNotExist(err) {
		err := os.MkdirAll(cfg.MediaFolder, os.ModePerm)
		if err != nil {
			logger.Fatalw("not possible to create media folder directory", "err", err)
		}
		logger.Info("media folder was created", "path", cfg.MediaFolder)
	}

	fmt.Println(cfg)

	// database
	db, err := database.NewConnection(
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

	// migrations
	driver, err := postgresmigrate.WithInstance(db, &postgresmigrate.Config{})
	if err != nil {
		logger.Fatalw("could not create driver for migrations", "error", err)
	}
	migrator, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		logger.Fatalf("could not create database instance for migrations", "error", err)
	}
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Fatalw("could not run migrations up", "err", err)
	}

	// store
	store := postgres.NewPostgresStore(db)

	// cache
	var rdb *redis.Client
	if cfg.Cacher.IsEnable {
		rdb = redisstore.NewRedisClient(
			cfg.Cacher.Redis.Addr,
			cfg.Cacher.Redis.Password,
			cfg.Cacher.Redis.Db,
		)
		logger.Infow("redis cache connection established", "addr", cfg.Cacher.Redis.Addr)
	}
	cacheStore := redisstore.NewRedisStore(rdb)

	// smtp
	smtpServer := env.GetString("SMTP_SERVER", "gmail")
	fromEmail := env.GetString("FROM_EMAIL", "email")
	emailPassword := env.GetString("EMAIL_PASSWORD", "password")
	clientMailer, err := mailer.NewGomail(smtpServer, fromEmail, emailPassword)
	if err != nil {
		logger.Panicw("could not create mailer", "error", err)
	}

	// ratelimiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.RateLimiter.RequestPerTimeFrame,
		cfg.RateLimiter.TimeFrame,
	)

	// services config
	serviceCfg := config.ServiceCfg{
		Logger:     logger,
		Store:      store,
		Cfg:        cfg,
		Mailer:     clientMailer,
		CacheStore: cacheStore,
	}
	services := service.NewServices(&serviceCfg)

	// oauth
	key := env.GetString("SESSION_SECRET", "")
	cookieStore := sessions.NewCookieStore([]byte(key))
	cookieStore.MaxAge(60 * 10)
	cookieStore.Options.Secure = true
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.SameSite = http.SameSiteLaxMode
	gothic.Store = cookieStore
	goth.UseProviders(
		google.New(
			cfg.OAuth.Google.Key,
			cfg.OAuth.Google.Secret,
			cfg.OAuth.Google.CallbackURI,
			"email", "profile",
		),
	)

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

	cronCtx, cronCancel := context.WithCancel(context.Background())
	cronjobs.PurgeUnconfirmedUsers(cronCtx, store, 1*time.Minute, logger)

	mux := app.Mount()
	if err := app.Run(mux); err != nil {
		app.Logger.Fatal(err)
	}

	cronCancel()
}
