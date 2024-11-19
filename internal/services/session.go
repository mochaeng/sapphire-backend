package services

import (
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"go.uber.org/zap"
)


type SessionService struct {
	store  *store.Store
	cfg    *config.Cfg
	logger *zap.SugaredLogger
}
