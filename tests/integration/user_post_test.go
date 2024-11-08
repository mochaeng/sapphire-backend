package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mochaeng/sapphire-backend/internal/app"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserPosterFlowSuite struct {
	suite.Suite
	pgContainer    *testutils.PostgresTestContainer
	redisContainer *testutils.RedisTestContainer
	ctx            context.Context
	db             *sql.DB
	app            *app.Application
	mux            http.Handler
}

func (suite *UserPosterFlowSuite) SetupSuite() {
	suite.ctx = context.Background()

	pgContainer, err := testutils.CreatePostgresContainer(suite.ctx)
	require.NoError(suite.T(), err, "could not create posgres container, error: %s", err)
	db := testutils.NewPostgresConnection(pgContainer.ConnString)
	suite.db = db
	suite.pgContainer = pgContainer

	redisContainer, err := testutils.CreateRedisContainer(suite.ctx)
	require.NoError(suite.T(), err, "could not create redis container, error: %s", err)
	suite.redisContainer = redisContainer

	app, err := createNewAppSuite(db, redisContainer)
	require.NoError(suite.T(), err, "could not create app, error: %s", err)

	suite.app = app
	suite.mux = app.Mount()

	driver, err := postgres.WithInstance(suite.db, &postgres.Config{})
	require.NoError(suite.T(), err, "could not create driver, error: %s", err)
	migrator, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	require.NoError(suite.T(), err, "could not apply migrations, error: %s", err)

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		suite.T().Fatalf("could not apply migrations, error: %s", err)
	}
	err = testutils.RunTestSeed(suite.db, integrationSeedPath)
	require.NoError(suite.T(), err, "could not seed test database")
}

func (suite *UserPosterFlowSuite) TearDownSuite() {
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

func (suite *UserPosterFlowSuite) TestMainFlow() {
	userPayload := &models.RegisterUserPayload{
		Username:  "gaga77",
		FirstName: "Lady",
		LastName:  "Gaga",
		Email:     "gaga@jype.com",
		Password:  "password123",
	}
	response := suite.registerUser(userPayload)
	suite.activateUser(response.Token)
	_ = suite.generateUserToken(userPayload.Email, userPayload.Password)
}

func TestUserPosterFlowSuite(t *testing.T) {
	suite.Run(t, new(UserPosterFlowSuite))
}

func (suite *UserPosterFlowSuite) registerUser(payload *models.RegisterUserPayload) *models.RegisterUserResponse {
	t := suite.T()
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, "could not marshal user struct")

	req, err := http.NewRequest(http.MethodPost, "/v1/auth/register/user", bytes.NewReader(jsonData))
	require.NoError(t, err, "could not make HTTP Post request to register router, error: %s", err)

	rr := testutils.ExecuteRequest(req, suite.mux)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response struct {
		Data models.RegisterUserResponse `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.Equal(t, response.Data.Username, payload.Username)
	assert.Equal(t, response.Data.IsActive, false)
	return &response.Data
}

func (suite *UserPosterFlowSuite) activateUser(token string) {
	t := suite.T()
	activationURL := fmt.Sprintf("/v1/user/activate/%s", token)
	req, err := http.NewRequest(http.MethodPut, activationURL, nil)
	require.NoError(t, err, "could not make put resquest to activate user router, error: %s", err)
	rr := testutils.ExecuteRequest(req, suite.mux)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func (suite *UserPosterFlowSuite) generateUserToken(email string, password string) string {
	t := suite.T()
	payload := models.CreateUserTokenPayload{
		Email:    email,
		Password: password,
	}
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, "could not marshal data")
	req, err := http.NewRequest(http.MethodPost, "/v1/auth/token", bytes.NewReader(jsonData))
	require.NoError(t, err, "could not make post request to generate token")
	rr := testutils.ExecuteRequest(req, suite.mux)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response struct {
		Data models.CreateTokenResponse `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err, "could not parse response")
	return response.Data.Token
}
