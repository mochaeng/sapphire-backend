package integration

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mochaeng/sapphire-backend/internal/app"
	"github.com/mochaeng/sapphire-backend/internal/models/payloads"
	"github.com/mochaeng/sapphire-backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	signinRouter   = "/v1/auth/signin/"
	signupRouter   = "/v1/auth/signup/"
	activateRouter = "/v1/verify-email/"
)

type MainFlowsSuite struct {
	suite.Suite
	pgContainer    *testutils.PostgresTestContainer
	redisContainer *testutils.RedisTestContainer
	ctx            context.Context
	db             *sql.DB
	app            *app.Application
	mux            http.Handler
}

func (suite *MainFlowsSuite) TestUserCreationAndPosting() {
	t := suite.T()
	userPayload := &payloads.RegisterUserPayload{
		Username:  "gaga77",
		FirstName: "Lady",
		Email:     "gaga@jype.com",
		Password:  "password123",
	}
	response := registerUser(t, suite.mux, userPayload)
	assert.Equal(t, response.Data.Username, userPayload.Username)
	assert.Equal(t, response.Data.IsActive, false)

	activateUser(t, suite.mux, response.Data.Token)
	responseWithCookie := signinUser(t, suite.mux, userPayload.Email, userPayload.Password)

	cookies := responseWithCookie.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	require.Equal(t, app.AuthTokenKey, cookie.Name)

	makePost(
		t,
		suite.mux,
		cookie,
		"tittle post",
		"this is a test using form-data",
		[]string{"integration", "test", "learning"},
	)
	makePost(
		t,
		suite.mux,
		cookie,
		"My second post",
		"Hallo everyone. Nice to meet yall",
		[]string{"first poster", "friendly"},
	)
}

func (suite *MainFlowsSuite) TestUserUniqueness() {
	t := suite.T()

	userPayload := &payloads.RegisterUserPayload{
		Username:  "chae77",
		FirstName: "son",
		Email:     "chaechae@jype.com",
		Password:  "password123",
	}
	response := registerUser(t, suite.mux, userPayload)
	assert.Equal(t, http.StatusCreated, response.rr.Code)
	assert.Equal(t, response.Data.Username, userPayload.Username)
	assert.Equal(t, response.Data.IsActive, false)

	userSameEmailPayload := &payloads.RegisterUserPayload{
		Username:  "aya",
		FirstName: "aya",
		Email:     "chaechae@jype.com",
		Password:  "password123",
	}
	response = registerUser(t, suite.mux, userSameEmailPayload)
	assert.Equal(t, http.StatusConflict, response.rr.Code)

	userSameUsernamePayload := &payloads.RegisterUserPayload{
		Username:  "eva_krauser",
		FirstName: "eva",
		Email:     "eva@jype.com",
		Password:  "password123",
	}
	response = registerUser(t, suite.mux, userSameUsernamePayload)
	assert.Equal(t, http.StatusConflict, response.rr.Code)
}

func TestUserPosterFlowSuite(t *testing.T) {
	suite.Run(t, new(MainFlowsSuite))
}

func (suite *MainFlowsSuite) SetupSuite() {
	suite.ctx = context.Background()

	pgContainer, err := testutils.CreatePostgresContainer(suite.ctx)
	require.NoError(suite.T(), err, ErrContainerNotStarting)
	db := testutils.NewPostgresConnection(pgContainer.ConnString)
	suite.db = db
	suite.pgContainer = pgContainer

	redisContainer, err := testutils.CreateRedisContainer(suite.ctx)
	require.NoError(suite.T(), err, ErrContainerNotStarting)
	suite.redisContainer = redisContainer

	parsedRedisConnStr := strings.TrimPrefix(redisContainer.ConnStrin, "redis://")
	app, err := createNewAppSuite(db, parsedRedisConnStr)
	require.NoError(suite.T(), err)

	suite.app = app
	suite.mux = app.Mount()

	driver, err := postgres.WithInstance(suite.db, &postgres.Config{})
	require.NoError(suite.T(), err, ErrMigrateDriverNotStarting)
	migrator, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	require.NoError(suite.T(), err, ErrMigrateApplying)

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		suite.T().Fatalf("%s, err: %s", ErrMigrateApplying.Error(), err)
	}
	err = testutils.RunTestSeed(suite.db, integrationSeedPath)
	require.NoError(suite.T(), err, ErrDatabaseSeed)
}

func (suite *MainFlowsSuite) TearDownSuite() {
	if err := suite.db.Close(); err != nil {
		log.Fatalf("could not close db connection, error: %s", err)
	}
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("could not terminating postgres container, error: %s", err)
	}
	if err := suite.redisContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("could not terminating redis container, error: %s", err)
	}
}
