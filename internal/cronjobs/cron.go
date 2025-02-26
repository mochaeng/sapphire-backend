package cronjobs

import (
	"context"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/store"
	"go.uber.org/zap"
)

func PurgeUnconfirmedUsers(ctx context.Context, s *store.Store, interval time.Duration, logger *zap.SugaredLogger) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.User.CleanUpExpiredPendingAccounts(ctx); err != nil {
					logger.Infow("unconfirmed users clean up failed", "err", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
