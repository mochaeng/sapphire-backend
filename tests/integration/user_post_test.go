package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
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
	jwtToken := suite.generateUserToken(userPayload.Email, userPayload.Password)

	suite.makePost(
		jwtToken,
		"tittle post",
		"this is a test using form-data",
		[]string{"integration", "test", "learning"},
	)
	suite.makePost(
		jwtToken,
		"My second post",
		"Hallo everyone. Nice to meet yall",
		[]string{"first poster", "friendly"},
	)
}

func TestUserPosterFlowSuite(t *testing.T) {
	suite.Run(t, new(UserPosterFlowSuite))
}

func (suite *UserPosterFlowSuite) registerUser(payload *models.RegisterUserPayload) *models.RegisterUserResponse {
	t := suite.T()
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, ErrPayloadMarshal)

	req, err := http.NewRequest(http.MethodPost, "/v1/auth/register/user", bytes.NewReader(jsonData))
	require.NoError(t, err, ErrRequestHTTP)

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
	require.NoError(t, err, ErrRequestHTTP)
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
	require.NoError(t, err, ErrPayloadMarshal)
	req, err := http.NewRequest(http.MethodPost, "/v1/auth/token", bytes.NewReader(jsonData))
	require.NoError(t, err, ErrRequestHTTP)
	rr := testutils.ExecuteRequest(req, suite.mux)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response struct {
		Data models.CreateTokenResponse `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err, ErrResponseParse)
	return response.Data.Token
}

func (suite *UserPosterFlowSuite) makePost(jwtToken, tittle, content string, tags []string) {
	t := suite.T()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("tittle", tittle)
	writer.WriteField("content", content)
	for _, tag := range tags {
		writer.WriteField("tags", tag)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/v1/post/", body)
	require.NoError(t, err, ErrRequestHTTP)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	rr := testutils.ExecuteRequest(req, suite.mux)
	var resp struct {
		Data models.CreatePostResponse `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rr.Code)
}
