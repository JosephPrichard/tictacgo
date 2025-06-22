package main

import (
	"TicTacGo/db"
	"fmt"

	//"TicTacGo/db"
	"context"
	_ "embed"
	"github.com/jackc/pgx/v5/pgxpool"

	//"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"

	"TicTacGo/service"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	//"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

// var testDbUser = "postgres"
// var testDbName = "tictacgo"
// var testDbPass = "password123"
// var testDbPort = 9876
var testToken = uuid.New().String()

//func CreateEmbeddedDb(t *testing.T) *embeddedpostgres.EmbeddedPostgres {
//	config := embeddedpostgres.DefaultConfig().
//		Username(testDbUser).
//		Password(testDbPass).
//		Database(testDbName).
//		Version(embeddedpostgres.V15).
//		RuntimePath("/tmp").
//		Port(uint32(testDbPort)).
//		StartTimeout(30 * time.Second)
//	t.Logf("starting the embedded test database with config: %v", config)
//
//	postgres := embeddedpostgres.NewDatabase(config)
//	err := postgres.Start()
//	if err != nil {
//		t.Fatalf("failed to start embedded datanase: %v", err)
//	}
//
//	t.Logf("finished setting up the embedded test database")
//
//	return postgres
//}
//
//func CreateTestDatabase(t *testing.T) {
//	ctx := context.Background()
//
//	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s port=%d", testDbUser, testDbName, testDbPass, testDbPort))
//	if err != nil {
//		t.Fatalf("failed to connect to database with err: %v", err)
//	}
//	defer pool.Close()
//
//	conn, err := pool.Acquire(ctx)
//	if err != nil {
//		t.Fatalf("failed to acquire a db conn with err: %v", err)
//	}
//	defer conn.Release()
//
//	_, err = conn.Exec(ctx, db.CreateSchema)
//	if err != nil {
//		t.Fatalf("failed to execute a conn with err: %v", err)
//	}
//
//	t.Logf("successfully created the database schema")
//}

func TestServer(t *testing.T) {
	postgres := CreateEmbeddedDb(t)
	defer func(postgres *embeddedpostgres.EmbeddedPostgres) {
		err := postgres.Stop()
		if err != nil {
			t.Fatalf("failed to stop test db with err: %v", err)
		}
	}(postgres)

	CreateTestDatabase(t)

	go StartGrpcServer(ServerConfig{
		port:       8080,
		dbUser:     testDbUser,
		dbName:     testDbUser,
		dbPassword: testDbPass,
		dbPort:     testDbPort,
	})

	TestCreateGame(t)
}

func TestCreateGame(t *testing.T) {
	t.Logf("starting TestCreateGame")

	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("did not connect to TicTacGoService server: %v", err)
	}
	client := service.NewTicTacGoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	in := &service.CreateGameReq{
		Token: &service.AuthToken{Token: testToken},
	}

	game, err := client.CreateGame(ctx, in)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	t.Logf("received game: %v from CreateGame call", game)
}
