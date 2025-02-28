package config

import (
	"time"

	"github.com/mochaeng/sapphire-backend/internal/mailer"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"github.com/mochaeng/sapphire-backend/internal/store/cache"
	"go.uber.org/zap"
)

const MaxMediaUploadSize int64 = 10_485_760 // 10MB

type roleHelper struct {
	ID    int
	Level int
}

var (
	Roles = map[string]roleHelper{
		"user": {
			ID:    1,
			Level: 1,
		},
		"moderator": {
			ID:    2,
			Level: 2,
		},
		"admin": {
			ID:    3,
			Level: 3,
		},
	}
)

type Cfg struct {
	Addr        string
	AppName     string
	DbConfig    DbCfg
	Version     string
	MediaFolder string
	// ApiURL      string
	Mail        MailCfg
	Auth        AuthCfg
	Cacher      CacheCfg
	RateLimiter RateLimiterConfig
	Env         string
	FrontedURL  string
	ApiBasePath string
	OAuth       OAuthConfig
}

type DbCfg struct {
	Addr               string
	MaxOpenConns       int
	MaxIdleConns       int
	MaxConnIdleSeconds int
}

type ServiceCfg struct {
	Logger     *zap.SugaredLogger
	Store      *store.Store
	CacheStore *cache.Store
	Cfg        *Cfg
	Mailer     mailer.Client
}

type CacheCfg struct {
	Redis    RedisCfg
	IsEnable bool
}

type RedisCfg struct {
	Addr     string
	Password string
	Db       int
}

type AuthCfg struct {
	Basic BasicAuthCfg
	Token TokenCfg
}

type TokenCfg struct {
	Secret  string
	Expired time.Duration
	Issuer  string
}

type BasicAuthCfg struct {
	Username string
	Password string
}

type MailCfg struct {
	Expired   time.Duration
	FromEmail string
}

type RateLimiterConfig struct {
	RequestPerTimeFrame int
	TimeFrame           time.Duration
	IsEnable            bool
}

type OAuthConfig struct {
	Google GoogleOAuth
}

type GoogleOAuth struct {
	Key         string
	Secret      string
	CallbackURI string
}
