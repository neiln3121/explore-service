package storage_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neiln3121/explore-service/database/migrations"
	"github.com/neiln3121/explore-service/internal/storage"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type StorageSuite struct {
	suite.Suite
	container testcontainers.Container
	repo      *storage.Storage
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}

func (s *StorageSuite) SetupSuite() {
	ctx := context.Background()

	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("postgres"),
	)
	s.Require().NoError(err)

	dbURL, err := ctr.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// run migrations + seeds
	mg := migrations.GetMigrationSource()
	migrate.SetTable("migrations")

	_, err = migrate.Exec(db, "postgres", mg, migrate.Up)
	if err != nil {
		log.Fatal(fmt.Errorf("migrations failed: %w", err))
	}

	s.container = ctr
	s.repo = storage.New(db)
}

func (s *StorageSuite) TearDownSuite() {
	err := s.repo.Close()
	s.Require().NoError(err)
	testcontainers.CleanupContainer(s.T(), s.container)
}

func (s *StorageSuite) TestGetLikedDecisionsCount() {
	ctx := context.Background()

	res, err := s.repo.GetLikedDecisionsCount(ctx, "user-1", true)
	s.Require().NoError(err)
	s.Assert().Equal(3, res)
}

func (s *StorageSuite) TestGetLikedDecision() {
	ctx := context.Background()

	res, err := s.repo.GetLikedDecision(ctx, "user-1", "user-2")
	s.Require().NoError(err)
	s.Assert().Equal(true, res)

	res, err = s.repo.GetLikedDecision(ctx, "user-1", "user-3")
	s.Require().NoError(err)
	s.Assert().Equal(false, res)
}

func (s *StorageSuite) TestGetLikedDecisions() {
	ctx := context.Background()

	res, err := s.repo.GetLikedDecisions(ctx, "user-1", true, nil, nil)
	s.Require().NoError(err)
	s.Assert().Len(res, 3)

	res, err = s.repo.GetLikedDecisions(ctx, "user-1", false, nil, nil)
	s.Require().NoError(err)
	s.Assert().Len(res, 1)

	// Pagination tests
	limit := uint32(2)
	res, err = s.repo.GetLikedDecisions(ctx, "user-1", true, nil, &limit)
	s.Require().NoError(err)
	s.Assert().Len(res, 2)

	nextToken := &res[len(res)-1].ID
	res, err = s.repo.GetLikedDecisions(ctx, "user-1", true, nextToken, nil)
	s.Require().NoError(err)
	s.Assert().Len(res, 1)
}

func (s *StorageSuite) Test_PutDecision() {
	ctx := context.Background()

	err := s.repo.PutDecision(ctx, "user-2", "user-3", true)
	s.Require().NoError(err)

	err = s.repo.PutDecision(ctx, "user-3", "user-4", false)
	s.Require().NoError(err)
}

func (s *StorageSuite) Test_GetNewLikedDecisions() {
	ctx := context.Background()

	res, err := s.repo.GetNewLikedDecisions(ctx, "user-1", true, nil, nil)
	s.Require().NoError(err)
	s.Assert().Len(res, 3)

	// User 5 is one of the 3 new likes
	s.Assert().Equal(res[0].ActorID, "user-5")

	// User 5 likes user 1 in return
	err = s.repo.PutMutualDecisions(ctx, "user-5", "user-1", true, true)
	s.Require().NoError(err)

	// Now only 2 new likes
	res, err = s.repo.GetNewLikedDecisions(ctx, "user-1", true, nil, nil)
	s.Require().NoError(err)
	s.Assert().Len(res, 2)

	// Pagination tests
	limit := uint32(1)
	res, err = s.repo.GetNewLikedDecisions(ctx, "user-1", true, nil, &limit)
	s.Require().NoError(err)
	s.Assert().Len(res, 1)

	nextToken := &res[len(res)-1].ID
	res, err = s.repo.GetNewLikedDecisions(ctx, "user-1", true, nextToken, nil)
	s.Require().NoError(err)
	s.Assert().Len(res, 1)
}
