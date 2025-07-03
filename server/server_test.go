package server

import (
	"TicTacGo/db"
	"TicTacGo/tictactoe"
	"context"
	_ "embed"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/testing/protocmp"
	"io"
	"log"
	"net"
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
		BoardState: "x_o______",
		XTurn:      true,
		Result:     tictactoe.Playing,
		Steps: []*pb.Step{
			{GameId: 1, Ord: 0, XTurn: false, Board: "x________", Result: tictactoe.Playing, MoveRow: 0, MoveCol: 0},
			{GameId: 1, Ord: 1, XTurn: true, Board: "x_o______", Result: tictactoe.Playing, MoveRow: 0, MoveCol: 2},
		},
	},
	{
		Id:         2,
		XPlayer:    &pb.Player{Id: 2, Username: "user2"},
		OPlayer:    &pb.Player{Id: 1, Username: "user1"},
		BoardState: "_________",
		XTurn:      true,
		Result:     tictactoe.Playing,
		Steps:      []*pb.Step{},
	},
	{
		Id:         3,
		XPlayer:    &pb.Player{Id: 1, Username: "user1"},
		OPlayer:    &pb.Player{Id: 3, Username: "user3"},
		BoardState: "_________",
		XTurn:      true,
		Result:     tictactoe.Forfeit,
		Steps: []*pb.Step{
			{GameId: 3, Ord: 0, XTurn: true, Board: "_________", Result: tictactoe.Forfeit},
		},
	},
	{
		Id:         4,
		XPlayer:    &pb.Player{Id: 1, Username: "user1"},
		BoardState: "_________",
		XTurn:      true,
		Result:     tictactoe.Playing,
		Steps:      []*pb.Step{},
	},
}

type TestArgs struct {
	client  pb.TicTacGoServiceClient
	pool    *pgxpool.Pool
	queries *db.Queries
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
	t.Logf("starting the embedded test database with config.go: %v", config)

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

func seedTestData(c TestArgs) {
	ctx := context.Background()

	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("failed to acquire a db conn with err: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS player_accounts, player_sessions, games, game_steps;")
	if err != nil {
		log.Fatalf("failed to drop schema with err: %v", err)
	}

	_, err = conn.Exec(ctx, db.CreateSchema)
	if err != nil {
		log.Fatalf("failed to execute CreateSchema with err: %v", err)
	}

	_, err = conn.Exec(ctx, db.SeedTestData)
	if err != nil {
		log.Fatalf("failed to execute SeedTestData with err: %v", err)
	}

	log.Printf("successfully created the database schema")
}

func serve(ctx context.Context, t *testing.T, pool *pgxpool.Pool) (pb.TicTacGoServiceClient, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	server := &GrpcServer{Queries: db.New(pool), Pool: pool}

	baseServer := grpc.NewServer()
	pb.RegisterTicTacGoServiceServer(baseServer, server)
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			t.Logf("error serving server: %v", err)
		}
	}()

	closer := func() {
		err := lis.Close()
		if err != nil {
			t.Logf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	dial := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dial), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Logf("error connecting to server: %v", err)
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

	queries := db.New(pool)

	client, closer := serve(ctx, t, pool)
	defer closer()

	args := TestArgs{
		pool:    pool,
		client:  client,
		queries: queries,
	}

	t.Run("RegisterAndLogin", func(t *testing.T) {
		testRegisterAndLogin(t, args)
	})
	t.Run("GetPlayers", func(t *testing.T) {
		testGetPlayers(t, args)
	})
	t.Run("CreateGame", func(t *testing.T) {
		testCreateGame(t, args)
	})
	t.Run("GetGame", func(t *testing.T) {
		testGetGame(t, args)
	})
	t.Run("GetGames", func(t *testing.T) {
		testGetGames(t, args)
	})
	t.Run("ListenSteps", func(t *testing.T) {
		testListenSteps(t, args)
	})
	t.Run("MakeMove", func(t *testing.T) {
		testMakeMove(t, args)
	})
}

func testRegisterAndLogin(t *testing.T, args TestArgs) {
	seedTestData(args)

	type RegisterTest struct {
		username string
		password string
		expId    int64
		expCode  codes.Code
	}

	registerTests := []RegisterTest{
		{username: "user6", password: "password123", expId: 6, expCode: 0},
		{username: "user3", password: "password123", expCode: codes.AlreadyExists},
	}

	for i, test := range registerTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			in := &pb.CredentialsReq{
				Username: test.username,
				Password: test.password,
			}

			player, err := args.client.Register(ctx, in)
			if test.expCode == 0 {
				assert.Nil(t, err)

				assert.Equal(t, test.expId, player.Id)
				assert.Equal(t, test.username, player.Username)

				dbPlayer, err := args.queries.GetPlayer(ctx, 6)
				if err != nil {
					t.Fatalf("failed to get game for assert: %v", err)
				}

				assert.Equal(t, test.expId, dbPlayer.ID)
				assert.Equal(t, test.username, dbPlayer.Username)
			}
			if test.expCode != 0 {
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expCode, s.Code())
			}
		})
	}

	type LoginTest struct {
		username string
		password string
		expCode  codes.Code
	}

	loginTests := []LoginTest{
		{username: "user6", password: "password123", expCode: 0},
		{username: "user6", password: "password-incorrect", expCode: codes.PermissionDenied},
		{username: "user10", password: "password123", expCode: codes.PermissionDenied},
	}

	for i, test := range loginTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			in := &pb.CredentialsReq{
				Username: test.username,
				Password: test.password,
			}

			_, err := args.client.Login(ctx, in)
			if test.expCode == 0 {
				assert.Nil(t, err)
			}
			if test.expCode != 0 {
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expCode, s.Code())
			}
		})
	}
}

func testGetPlayers(t *testing.T, args TestArgs) {
	seedTestData(args)

	type Test struct {
		in         *pb.GetPlayersReq
		expPlayers []*pb.Player
	}

	tests := []Test{
		{
			in: &pb.GetPlayersReq{Page: 1, PerPage: 3},
			expPlayers: []*pb.Player{
				{Username: "user1", Id: 1, Cnt: 1},
				{Username: "user2", Id: 2, Cnt: 0},
				{Username: "user3", Id: 3, Cnt: 1},
			},
		},
		{
			in: &pb.GetPlayersReq{Page: 2, PerPage: 3},
			expPlayers: []*pb.Player{
				{Username: "user4", Id: 4},
				{Username: "user5", Id: 5},
			},
		},
		{
			in:         &pb.GetPlayersReq{Page: 5, PerPage: 20},
			expPlayers: nil,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			players, err := args.client.GetPlayers(ctx, test.in)
			if err != nil {
				t.Fatalf("failed to get players: %v", err)
			}

			diff := cmp.Diff(test.expPlayers, players.Players, protocmp.Transform())
			assert.Equal(t, "", diff)
		})
	}
}

func testCreateGame(t *testing.T, args TestArgs) {
	seedTestData(args)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	in := &pb.CreateGameReq{Token: "User1Token"}

	game, err := args.client.CreateGame(ctx, in)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	expectedGame := &pb.Game{
		Id: 5,
		XPlayer: &pb.Player{
			Id:       1,
			Username: "user1",
		},
		BoardState: "_________",
		XTurn:      true,
		Result:     tictactoe.Playing,
		Steps:      []*pb.Step{},
	}
	diff := cmp.Diff(expectedGame, game, protocmp.Transform(), protocmp.IgnoreFields(expectedGame, "updated_on", "started_on"))
	assert.Equal(t, "", diff)

	dbGame, err := args.queries.GetGame(ctx, 5)
	if err != nil {
		t.Fatalf("failed to get game for assert: %v", err)
	}

	expectedDbGame := db.GetGameRow{
		ID:          5,
		XPlayer:     1,
		OPlayer:     pgtype.Int8{},
		BoardState:  "_________",
		XTurn:       pgtype.Bool{Bool: true, Valid: true},
		Result:      tictactoe.Playing,
		XPlayerName: pgtype.Text{String: "user1", Valid: true},
		OPlayerName: pgtype.Text{String: "", Valid: false},
	}

	diff = cmp.Diff(expectedDbGame, dbGame, cmpopts.IgnoreFields(dbGame, "UpdatedOn", "StartedOn"))
	assert.Equal(t, "", diff)
}

func testGetGame(t *testing.T, args TestArgs) {
	seedTestData(args)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	in := &pb.GetGameReq{Id: 1}

	game, err := args.client.GetGame(ctx, in)
	if err != nil {
		t.Fatalf("failed to get game: %v", err)
	}

	expectedGame := &testPbGames[0]

	diff := cmp.Diff(expectedGame, game, protocmp.Transform(), protocmp.IgnoreFields(&pb.Game{}, "updated_on", "started_on"))
	assert.Equal(t, "", diff)
}

func testGetGames(t *testing.T, args TestArgs) {
	seedTestData(args)

	type Test struct {
		params   *pb.GetGamesReq
		expGames []*pb.Game
	}

	tests := []Test{
		{
			params:   &pb.GetGamesReq{Page: 1, PerPage: 20},
			expGames: []*pb.Game{&testPbGames[0], &testPbGames[1], &testPbGames[2], &testPbGames[3]},
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

			games, err := args.client.GetGames(ctx, test.params)
			if err != nil {
				t.Fatalf("failed to get games: %v", err)
			}

			expectedGames := &pb.Games{Games: test.expGames}

			diff := cmp.Diff(expectedGames, games, protocmp.Transform(), protocmp.IgnoreFields(&pb.Game{}, "updated_on", "started_on"))
			assert.Equal(t, "", diff)
		})
	}
}

func testListenSteps(t *testing.T, args TestArgs) {
	seedTestData(args)

	type Test struct {
		id       int64
		expSteps []*pb.Step
	}

	step1 := &pb.Step{GameId: 1, Ord: 1, XTurn: true, Board: "x_o______", Result: tictactoe.Playing, MoveRow: 0, MoveCol: 2}
	step2 := &pb.Step{GameId: 3, Ord: 0, XTurn: true, Board: "_________", Result: tictactoe.Forfeit}
	tests := []Test{
		{
			id:       1,
			expSteps: []*pb.Step{step1, step1},
		},
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

			stream, err := args.client.ListenSteps(ctx, in)
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

			diff := cmp.Diff(test.expSteps, steps, protocmp.Transform())
			assert.Equal(t, "", diff)
		})
	}
}

func testMakeMove(t *testing.T, args TestArgs) {
	seedTestData(args)

	type Test struct {
		in        *pb.MakeMoveReq
		expGame   *pb.Game
		expDbGame *db.GetGameRow
		expCode   codes.Code
	}

	tests := []Test{
		{
			in:      &pb.MakeMoveReq{GameId: 1, Row: 0, Col: 0, Token: "User1Token"}, // illegal move
			expCode: codes.InvalidArgument,
		},
		{
			in:      &pb.MakeMoveReq{GameId: 1, Row: 0, Col: 0, Token: "InvalidToken"}, // invalid token
			expCode: codes.PermissionDenied,
		},
		{
			in:      &pb.MakeMoveReq{GameId: 2, Row: 0, Col: 0, Token: "User1Token"}, // not player's turn
			expCode: codes.PermissionDenied,
		},
		{
			in:      &pb.MakeMoveReq{GameId: 3, Row: 0, Col: 0, Token: "User1Token"}, // cannot make move on game not in play
			expCode: codes.PermissionDenied,
		},
		{
			in:      &pb.MakeMoveReq{GameId: 4, Row: 0, Col: 0, Token: "User1Token"}, // cannot make move on game that is not started
			expCode: codes.PermissionDenied,
		},
		{
			in: &pb.MakeMoveReq{GameId: 1, Row: 0, Col: 1, Token: "User1Token"}, // success case
			expGame: &pb.Game{
				Id:         1,
				XPlayer:    &pb.Player{Id: 1, Username: "user1"},
				OPlayer:    &pb.Player{Id: 2, Username: "user2"},
				BoardState: "xxo______",
				XTurn:      false,
				Result:     tictactoe.Playing,
				Steps:      []*pb.Step{},
			},
			expDbGame: &db.GetGameRow{
				ID:          1,
				XPlayer:     1,
				OPlayer:     pgtype.Int8{Int64: 2, Valid: true},
				BoardState:  "xxo______",
				XTurn:       pgtype.Bool{Bool: false, Valid: true},
				Result:      tictactoe.Playing,
				XPlayerName: pgtype.Text{String: "user1", Valid: true},
				OPlayerName: pgtype.Text{String: "user2", Valid: true},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			game, err := args.client.MakeMove(ctx, test.in)
			if test.expCode == 0 {
				assert.Nil(t, err)

				diff := cmp.Diff(test.expGame, game, protocmp.Transform(), protocmp.IgnoreFields(&pb.Game{}, "updated_on", "started_on"))
				assert.Equal(t, "", diff)
			}
			if test.expCode != 0 {
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expCode, s.Code())
			}

			if test.expDbGame != nil {
				dbGame, err := args.queries.GetGame(ctx, test.in.GameId)
				if err != nil {
					t.Fatalf("failed to get game for assert: %v", err)
				}

				diff := cmp.Diff(*test.expDbGame, dbGame, cmpopts.IgnoreFields(dbGame, "UpdatedOn", "StartedOn"))
				assert.Equal(t, "", diff)
			}
		})
	}
}
