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
	"net/http/httptest"
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

const (
	signinRouter   = "/v1/auth/signin/"
	signupRouter   = "/v1/auth/signup/"
	activateRouter = "/v1/verify-email/"
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

func (suite *UserPosterFlowSuite) TestUserCreationAndPosting() {
	t := suite.T()
	userPayload := &models.RegisterUserPayload{
		Username:  "gaga77",
		FirstName: "Lady",
		Email:     "gaga@jype.com",
		Password:  "password123",
	}
	response := suite.registerUser(userPayload)
	assert.Equal(t, response.Data.Username, userPayload.Username)
	assert.Equal(t, response.Data.IsActive, false)

	suite.activateUser(response.Data.Token)
	responseWithCookie := suite.signinUser(userPayload.Email, userPayload.Password)

	cookies := responseWithCookie.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	require.Equal(t, app.AuthTokenKey, cookie.Name)

	suite.makePost(
		cookie,
		"tittle post",
		"this is a test using form-data",
		[]string{"integration", "test", "learning"},
	)
	suite.makePost(
		cookie,
		"My second post",
		"Hallo everyone. Nice to meet yall",
		[]string{"first poster", "friendly"},
	)
}

func (suite *UserPosterFlowSuite) TestUserUniqueness() {
	t := suite.T()

	userPayload := &models.RegisterUserPayload{
		Username:  "chae77",
		FirstName: "son",
		Email:     "chaechae@jype.com",
		Password:  "password123",
	}
	response := suite.registerUser(userPayload)
	assert.Equal(t, http.StatusCreated, response.rr.Code)
	assert.Equal(t, response.Data.Username, userPayload.Username)
	assert.Equal(t, response.Data.IsActive, false)

	userSameEmailPayload := &models.RegisterUserPayload{
		Username:  "aya",
		FirstName: "aya",
		Email:     "chaechae@jype.com",
		Password:  "password123",
	}
	response = suite.registerUser(userSameEmailPayload)
	assert.Equal(t, http.StatusConflict, response.rr.Code)

	userSameUsernamePayload := &models.RegisterUserPayload{
		Username:  "eva_krauser",
		FirstName: "eva",
		Email:     "eva@jype.com",
		Password:  "password123",
	}
	response = suite.registerUser(userSameUsernamePayload)
	assert.Equal(t, http.StatusConflict, response.rr.Code)
}

func TestUserPosterFlowSuite(t *testing.T) {
	suite.Run(t, new(UserPosterFlowSuite))
}

func (suite *UserPosterFlowSuite) registerUser(payload *models.RegisterUserPayload) RegisterUserResponse {
	t := suite.T()
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, ErrPayloadMarshal)

	req, err := http.NewRequest(http.MethodPost, signupRouter, bytes.NewReader(jsonData))
	require.NoError(t, err, ErrRequestHTTP)

	rr := testutils.ExecuteRequest(req, suite.mux)

	response := RegisterUserResponse{rr: rr}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	return response
}

func (suite *UserPosterFlowSuite) activateUser(token string) {
	t := suite.T()
	activationURL := fmt.Sprintf("%s%s", activateRouter, token)
	req, err := http.NewRequest(http.MethodPut, activationURL, nil)
	require.NoError(t, err, ErrRequestHTTP)
	rr := testutils.ExecuteRequest(req, suite.mux)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func (suite *UserPosterFlowSuite) signinUser(email string, password string) *httptest.ResponseRecorder {
	t := suite.T()
	payload := models.SigninPayload{
		Email:    email,
		Password: password,
	}
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, ErrPayloadMarshal)
	req, err := http.NewRequest(http.MethodPost, signinRouter, bytes.NewReader(jsonData))
	require.NoError(t, err, ErrRequestHTTP)
	rr := testutils.ExecuteRequest(req, suite.mux)
	assert.Equal(t, http.StatusCreated, rr.Code)
	return rr
}

func (suite *UserPosterFlowSuite) makePost(cookie *http.Cookie, tittle, content string, tags []string) {
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
	req.AddCookie(cookie)

	rr := testutils.ExecuteRequest(req, suite.mux)
	var resp struct {
		Data models.CreatePostResponse `json:"data"`
	}
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rr.Code)
}
