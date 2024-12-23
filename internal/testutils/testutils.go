package testutils

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresTestContainer struct {
	*postgres.PostgresContainer
	ConnString string
}

type RedisTestContainer struct {
	*redis.RedisContainer
	ConnStrin string
}

func NewPostgresConnection(connStr string) *sql.DB {
	maxOpenConns := 10
	maxIddleConns := 10
	maxConnIdleSeconds := 120
	db, err := database.NewConnection(
		connStr,
		maxOpenConns,
		maxIddleConns,
		maxConnIdleSeconds,
	)
	if err != nil {
		log.Fatalf("could not start database connection, err: %s", err)
	}
	return db
}

func CreateRedisContainer(ctx context.Context) (*RedisTestContainer, error) {
	redisContainer, err := redis.Run(ctx,
		"redis:6.2-alpine",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		// redis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}
	return &RedisTestContainer{
		RedisContainer: redisContainer,
		ConnStrin:      connStr,
	}, nil
}

func CreatePostgresContainer(ctx context.Context) (*PostgresTestContainer, error) {
	pgContainer, err := postgres.Run(ctx, "postgres:16.4",
		// postgres.WithInitScripts(filepath.Join("..", "testdata", "init-db.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}
	return &PostgresTestContainer{
		PostgresContainer: pgContainer,
		ConnString:        connStr,
	}, nil
}

func RunTestSeed(db *sql.DB, seedPath string) error {
	content, err := database.GetSeedContentFromFile(seedPath)
	if err != nil {
		return fmt.Errorf("could not read seed file, err: %s", err)
	}
	statements := bytes.Split(content, []byte(";"))
	for _, stmt := range statements {
		trimmedStmt := strings.TrimSpace(string(stmt))
		if len(trimmedStmt) > 0 {
			_, err := db.Exec(trimmedStmt)
			if err != nil {
				return fmt.Errorf("could not execute seed statement, err: %s", err)
			}
		}
	}
	return nil
}

func ExecuteRequest(r *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, r)
	return rr
}
