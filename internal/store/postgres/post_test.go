package postgres

import (
	"context"
	"log"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	postgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PostStoreTestSuite struct {
	suite.Suite
	pgContainer *testutils.PostgresTestContainer
	postStore   *PostStore
	ctx         context.Context
}

func (suite *PostStoreTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	pgContainer, err := testutils.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatalf("could not create postgres container, err: %s", err)
	}
	suite.pgContainer = pgContainer

	store := newTestPostStore(suite.pgContainer.ConnString)
	suite.postStore = store

	driver, err := postgres.WithInstance(suite.postStore.db, &postgres.Config{})
	require.NoError(suite.T(), err)

	migrator, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	require.NoError(suite.T(), err)

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		suite.T().Fatalf("could not apply up migrations, err: %s", err)
	}

	err = testutils.RunTestSeed(suite.postStore.db, unitSeedPath)
	require.NoError(suite.T(), err, "could not seed test database")
}

func (suite *PostStoreTestSuite) TearDownSuite() {
	if err := suite.postStore.db.Close(); err != nil {
		log.Fatalf("could not close db connection, error: %s", err)
	}
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("could not terminating postgres container, error: %s", err)
	}
}

func (suite *PostStoreTestSuite) TestCreatePost() {
	t := suite.T()
	post := models.Post{
		Tittle:  "my test post",
		Content: "this is a test post",
		User:    &models.User{ID: 1},
		Tags:    []string{"test"},
	}
	err := suite.postStore.Create(suite.ctx, &post)
	require.NoError(t, err, "could not add post")

	retrievedPost, err := suite.postStore.GetByID(suite.ctx, post.ID)
	require.NoError(t, err, "could not retrieve created post")
	assert.Equal(t, retrievedPost.Tittle, post.Tittle)
	assert.Equal(t, retrievedPost.Content, post.Content)
}

func (suite *PostStoreTestSuite) TestGetPost() {
	t := suite.T()
	post, err := suite.postStore.GetByID(suite.ctx, 1)
	require.NoError(t, err, "could not get post by id")
	assert.Equal(t, post.Tittle, "chaeyoung <3")
	assert.Equal(t, post.Content, "chaeyoung > lalisa")
}

func TestPostStoreSuite(t *testing.T) {
	suite.Run(t, new(PostStoreTestSuite))
}
