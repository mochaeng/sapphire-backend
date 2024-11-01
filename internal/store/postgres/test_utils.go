package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	migrationsPath = fmt.Sprintf("file://%s", filepath.Join("..", "..", "..", "migrate", "migrations"))
	seedPath       = filepath.Join("..", "..", "..", "migrate", "tests", "unit_seed.sql")
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnString string
}

func createDB(connStr string) *sql.DB {
	maxOpenConns := 10
	maxIddleConns := 10
	maxConnIdleSeconds := 120
	db, err := database.New(
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

func createPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
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
	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnString:        connStr,
	}, nil
}

func runTestSeed(db *sql.DB) error {
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
