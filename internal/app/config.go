package app

import (
	"time"

	"github.com/go-playground/validator/v10"
)

const MaxUploadSize int64 = 10_485_760 // 10MB

type roleHelper struct {
	id    int
	level int
}

var (
	Validate *validator.Validate
	roles    = map[string]roleHelper{
		"user": {
			id:    1,
			level: 1,
		},
		"moderator": {
			id:    2,
			level: 2,
		},
		"admin": {
			id:    3,
			level: 3,
		},
	}
)

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

type Cfg struct {
	Addr        string
	DbConfig    DbCfg
	Env         string
	Version     string
	MediaFolder string
	ApiURL      string
	FrontedURL  string
	Mail        MailCfg
	Auth        AuthCfg
	Cacher      CacheCfg
	RateLimiter RateLimiterConfig
}

type DbCfg struct {
	Addr               string
	MaxOpenConns       int
	MaxIdleConns       int
	MaxConnIdleSeconds int
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
