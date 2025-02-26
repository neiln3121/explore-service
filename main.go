package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"github.com/neiln3121/explore-service/database/migrations"
	contract "github.com/neiln3121/explore-service/explore"
	"github.com/neiln3121/explore-service/internal/api"
	"github.com/neiln3121/explore-service/internal/storage"
	migrate "github.com/rubenv/sql-migrate"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 3000, "the port for the server")
)

func main() {
	flag.Parse()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// run migrations + seeds
	mg := migrations.GetMigrationSource()
	migrate.SetTable("migrations")

	_, err = migrate.Exec(db, "postgres", mg, migrate.Up)
	if err != nil {
		log.Fatal(fmt.Errorf("migrations failed: %w", err))
	}

	repo := storage.New(db)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	s := grpc.NewServer()

	api := api.New(repo)
	contract.RegisterExploreAPIServer(s, api)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
