package main

import (
	"TicTacGo/db"
	"TicTacGo/tictactoe"
	"context"
	_ "embed"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/testing/protocmp"
	"io"
	"net"

	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"

	"TicTacGo/pb"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"google.golang.org/grpc"
)

var testDbUser = "postgres"
var testDbName = "tictacgo"
var testDbPass = "password123"
var testDbPort = 9876

var testPbGames = []pb.Game{
	{
		Id:         1,
		XPlayer:    &pb.Player{Id: 1, Username: "user1"},
		OPlayer:    &pb.Player{Id: 2, Username: "user2"},
		BoardState: []byte{1, 0, 2 /**/, 0, 0, 0 /**/, 0, 0, 0},
		XTurn:      true,
		Result:     tictactoe.Playing,
		Steps: []*pb.Step{
			{GameId: 1, Ord: 0, XTurn: false, Board: []byte{1, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0}, Result: tictactoe.Playing, MoveRow: 0, MoveCol: 0},
			{GameId: 1, Ord: 1, XTurn: true, Board: []byte{1, 0, 2 /**/, 0, 0, 0 /**/, 0, 0, 0}, Result: tictactoe.Playing, MoveRow: 0, MoveCol: 2},
		},
	},
	{
		Id:         2,
		XPlayer:    &pb.Player{Id: 2, Username: "user2"},
		OPlayer:    &pb.Player{Id: 1, Username: "user1"},
		BoardState: make([]byte, 9),
		XTurn:      true,
		Result:     tictactoe.Playing,
		Steps:      []*pb.Step{},
	},
	{
		Id:         3,
		XPlayer:    &pb.Player{Id: 1, Username: "user1"},
		OPlayer:    &pb.Player{Id: 3, Username: "user3"},
		BoardState: []byte{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		XTurn:      true,
		Result:     tictactoe.Forfeit,
		Steps: []*pb.Step{
			{GameId: 3, Ord: 0, XTurn: true, Board: []byte{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0}, Result: tictactoe.Forfeit},
		},
	},
}

func createEmbeddedDb(t *testing.T) func() {
	config := embeddedpostgres.DefaultConfig().
		Username(testDbUser).
		Password(testDbPass).
		Database(testDbName).
		Version(embeddedpostgres.V15).
		RuntimePath("/tmp").
		Port(uint32(testDbPort)).
		StartTimeout(30 * time.Second)
	t.Logf("starting the embedded test database with config: %v", config)

	postgres := embeddedpostgres.NewDatabase(config)
	err := postgres.Start()
	if err != nil {
		t.Fatalf("failed to start embedded datanase: %v", err)
	}

	t.Logf("finished setting up the embedded test database")

	closer := func() {
		err := postgres.Stop()
		if err != nil {
			t.Fatalf("failed to stop test db with err: %v", err)
		}
	}

	return closer
}

func seedTestData(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("failed to acquire a db conn with err: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS player_accounts, player_sessions, games, game_steps;")
	if err != nil {
		t.Fatalf("failed to drop schema with err: %v", err)
	}

	_, err = conn.Exec(ctx, db.CreateSchema)
	if err != nil {
		t.Fatalf("failed to execute CreateSchema with err: %v", err)
	}

	_, err = conn.Exec(ctx, db.SeedTestData)
	if err != nil {
		t.Fatalf("failed to execute SeedTestData with err: %v", err)
	}

	t.Logf("successfully created the database schema")
}

func serve(ctx context.Context, t *testing.T, pool *pgxpool.Pool) (pb.TicTacGoServiceClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	server := &GrpcServer{queries: db.New(pool), pool: pool}

	baseServer := grpc.NewServer()
	pb.RegisterTicTacGoServiceServer(baseServer, server)
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			t.Logf("error serving server: %v", err)
		}
	}()

	dial := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dial), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Logf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			t.Logf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := pb.NewTicTacGoServiceClient(conn)

	return client, closer
}

func TestServer(t *testing.T) {
	ctx := context.Background()

	closer := createEmbeddedDb(t)
	defer closer()

	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s port=%d", testDbUser, testDbName, testDbPass, testDbPort))
	if err != nil {
		t.Fatalf("failed to connect to database with err: %v", err)
	}
	defer pool.Close()

	client, closer := serve(ctx, t, pool)
	defer closer()

	t.Run("CreateGame", func(t *testing.T) {
		seedTestData(ctx, t, pool)
		testCreateGame(t, client)
	})
	t.Run("CreatePlayer", func(t *testing.T) {
		seedTestData(ctx, t, pool)
		testCreatePlayer(t, client)
	})
	t.Run("GetGame", func(t *testing.T) {
		seedTestData(ctx, t, pool)
		testGetGame(t, client)
	})
	t.Run("Login", func(t *testing.T) {
		seedTestData(ctx, t, pool)
		testLogin(t, client)
	})
	t.Run("Play", func(t *testing.T) {
		seedTestData(ctx, t, pool)
		testGetGames(t, client)
	})
	t.Run("ListenSteps", func(t *testing.T) {
		seedTestData(ctx, t, pool)
		testListenSteps(t, client)
	})
}

func testCreateGame(t *testing.T, client pb.TicTacGoServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	in := &pb.CreateGameReq{
		Token: &pb.AuthToken{Token: "User1Token"},
	}

	game, err := client.CreateGame(ctx, in)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	expectedGame := &pb.Game{
		Id: 4,
		XPlayer: &pb.Player{
			Id:       1,
			Username: "user1",
		},
		BoardState: make([]byte, 9),
		XTurn:      true,
		Result:     tictactoe.Playing,
		Steps:      []*pb.Step{},
	}

	if !cmp.Equal(game, expectedGame, protocmp.Transform(), protocmp.IgnoreFields(&pb.Game{}, "updated_on", "started_on")) {
		t.Fatalf("produced incorrrect result: expected: %+v, got: %+v", expectedGame, game)
	}
}

func testCreatePlayer(t *testing.T, client pb.TicTacGoServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	in := &pb.CredentialsReq{
		Username: "user4",
		Password: "password4",
	}

	player, err := client.CreatePlayer(ctx, in)
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	expectedPlayer := &pb.Player{
		Id:       4,
		Username: "user4",
	}

	if !cmp.Equal(player, expectedPlayer, protocmp.Transform()) {
		t.Fatalf("produced incorrrect result: expected: %+v, got: %+v", expectedPlayer, player)
	}
}

func testGetGame(t *testing.T, client pb.TicTacGoServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	in := &pb.GetGameReq{Id: 1}

	game, err := client.GetGame(ctx, in)
	if err != nil {
		t.Fatalf("failed to get game: %v", err)
	}

	expectedGame := &testPbGames[0]

	if !cmp.Equal(game, expectedGame, protocmp.Transform(), protocmp.IgnoreFields(&pb.Game{}, "updated_on", "started_on")) {
		t.Fatalf("produced incorrrect result: expected: %+v, got: %+v", expectedGame, game)
	}
}

func testLogin(t *testing.T, client pb.TicTacGoServiceClient) {
	type Test struct {
		username string
		password string
		expCode  codes.Code
	}

	tests := []Test{
		{username: "user1", password: "password123", expCode: 0},
		{username: "user1", password: "password-incorrect", expCode: codes.PermissionDenied},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			in := &pb.CredentialsReq{
				Username: test.username,
				Password: test.password,
			}

			_, err := client.Login(ctx, in)
			if test.expCode == 0 && err != nil {
				t.Fatal("produced an error, expected no error")
			}
			if test.expCode != 0 {
				if s, ok := status.FromError(err); !ok || s.Code() != test.expCode {
					t.Fatalf("produced incorrect error code, ok: %v, expected: %d, got: %d", ok, test.expCode, s.Code())
				}
			}
		})
	}
}

func testGetGames(t *testing.T, client pb.TicTacGoServiceClient) {
	type Test struct {
		params   *pb.GetGamesReq
		expGames []*pb.Game
	}

	tests := []Test{
		{
			params:   &pb.GetGamesReq{Page: 1, PerPage: 20},
			expGames: []*pb.Game{&testPbGames[0], &testPbGames[1], &testPbGames[2]},
		},
		{
			params:   &pb.GetGamesReq{Page: 1, OPlayer: &pb.Player{Id: 3}, PerPage: 20},
			expGames: []*pb.Game{&testPbGames[2]},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			games, err := client.GetGames(ctx, test.params)
			if err != nil {
				t.Fatalf("failed to get games: %v", err)
			}

			expectedGames := &pb.Games{Games: test.expGames}

			if !cmp.Equal(games, expectedGames, protocmp.Transform(), protocmp.IgnoreFields(&pb.Game{}, "updated_on", "started_on")) {
				t.Fatalf("produced incorrrect result: expected: %+v, got: %+v", expectedGames, games)
			}
		})
	}
}

func testListenSteps(t *testing.T, client pb.TicTacGoServiceClient) {
	type Test struct {
		id       int64
		expSteps []*pb.Step
	}

	step1 := &pb.Step{GameId: 1, Ord: 1, XTurn: true, Board: []byte{1, 0, 2 /**/, 0, 0, 0 /**/, 0, 0, 0}, Result: tictactoe.Playing, MoveRow: 0, MoveCol: 2}
	step2 := &pb.Step{GameId: 3, Ord: 0, XTurn: true, Board: []byte{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0}, Result: tictactoe.Forfeit}
	tests := []Test{
		{
			id:       1,
			expSteps: []*pb.Step{step1, step1},
		},
		//{
		//	id:      1,
		//	expSteps: []*pb.Step{},
		//},
		{
			id:       3,
			expSteps: []*pb.Step{step2},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			in := &pb.GetGameReq{Id: test.id}

			stream, err := client.ListenSteps(ctx, in)
			if err != nil {
				t.Fatalf("failed to open stream: %v", err)
			}

			var steps []*pb.Step

			for i := 0; i < 2; i++ {
				step, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("failed to recv on stream: %v", err)
				}
				steps = append(steps, step)
			}

			if !cmp.Equal(steps, test.expSteps, protocmp.Transform()) {
				t.Fatalf("produced incorrrect result: expected: %+v, got: %+v", test.expSteps, steps)
			}
		})
	}
}
