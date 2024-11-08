package postgres

import (
	"context"
	"log"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserStoreTestSuite struct {
	suite.Suite
	pgContainer *testutils.PostgresContainer
	userStore   *UserStore
	ctx         context.Context
}

func (suite *UserStoreTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testutils.CreatePostgresContainer(suite.ctx)
	suite.pgContainer = pgContainer

	store := newTestUserStore(suite.pgContainer.ConnString)
	suite.userStore = store

	driver, err := postgres.WithInstance(suite.userStore.db, &postgres.Config{})
	require.NoError(suite.T(), err)

	migrator, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	require.NoError(suite.T(), err)

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		suite.T().Fatalf("could not apply up migrations, err: %s", err)
	}

	err = testutils.RunTestSeed(suite.userStore.db, seedPath)
	require.NoError(suite.T(), err, "could no seed test database")
}

func (suite *UserStoreTestSuite) TearDownSuite() {
	if err := suite.userStore.db.Close(); err != nil {
		log.Fatalf("could not close db connection, error: %s", err)
	}
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("could not terminating postgres container, error: %s", err)
	}
}

func (suite *UserStoreTestSuite) TestCreateUser() {
	t := suite.T()
	user := &models.User{
		Username:  "tzuyyy",
		FirstName: "chou",
		LastName:  "tzuyu",
		Email:     "chou@jype.com",
		Role:      models.Role{ID: 1},
	}

	tx, err := suite.userStore.db.BeginTx(suite.ctx, nil)
	require.NoError(t, err, "could not begin transaction")

	err = suite.userStore.Create(suite.ctx, tx, user)
	require.NoError(t, err, "could not create user")

	err = tx.Commit()
	require.NoError(t, err, "could not commit transaction")

	retrivedUser, err := suite.userStore.GetByEmail(suite.ctx, "momo@mail.com")
	require.NoError(t, err, "could not retrieve user")
	assert.Equal(t, retrivedUser.FirstName, "momo")
	assert.Equal(t, retrivedUser.LastName, "hirai")
}

func TestUserStoreTestSuite(t *testing.T) {
	suite.Run(t, new(UserStoreTestSuite))
}
