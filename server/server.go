package main

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"TicTacGo/database"
	"TicTacGo/service"
	"TicTacGo/tictactoe"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const limit int32 = 20

type GrpcServer struct {
	service.UnimplementedTicTacGoServiceServer
	queries *database.Queries
	pool    *pgxpool.Pool
}

func (s *GrpcServer) CreateGame(ctx context.Context, in *service.CreateGameReq) (*service.Game, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	// get the session for the provided token
	token := string(in.GetToken().Token)

	log.Printf("called CreateGame with token: %v", token)

	sessRow, err := s.queries.GetSession(ctx, token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve session: %v", err)
	}

	// insert the newly constructed game
	timeNow := time.Now()
	board, turn := tictactoe.NewBoard()

	boardState := make([]byte, 9)
	copy(boardState, board[:])

	params := database.InsertGameParams{
		XPlayer:    sessRow.ID,
		BoardState: boardState,
		XTurn:      pgtype.Bool{Bool: turn, Valid: true},
		UpdatedOn:  pgtype.Timestamp{Time: timeNow, Valid: true},
		StartedOn:  pgtype.Timestamp{Time: timeNow, Valid: true},
	}

	gameId, err := s.queries.InsertGame(ctx, params)
	if err != nil {
		log.Fatalf("failed to insert game: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert game for params: %v", params)
	}

	log.Printf("inserted game with params: %v", params)

	// construct the game response for client
	game := service.Game{
		Id: gameId,
		XPlayer: &service.Player{
			Id:       sessRow.ID,
			Username: sessRow.Username,
		},
		BoardState: params.BoardState,
		XTurn:      params.XTurn.Bool,
		UpdatedOn:  &timestamppb.Timestamp{Seconds: int64(timeNow.Second())},
		StartedOn:  &timestamppb.Timestamp{Seconds: int64(timeNow.Second())},
		Steps:      []*service.Step{},
	}

	log.Printf("successfully created game: %v, board: %s", game.String(), tictactoe.BoardToString(board))

	return &game, nil
}

func (s *GrpcServer) CreatePlayer(ctx context.Context, in *service.CredentialsReq) (*service.Player, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	params := database.InsertPlayerParams{
		Username: in.Username,
		Passwd:   in.Password,
	}

	row, err := s.queries.InsertPlayer(ctx, params)
	if err != nil {
		log.Fatalf("failed to insert player: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert player for params: %s", params)
	}

	player := service.Player{
		Id:       row.ID,
		Username: row.Username,
	}

	log.Printf("successfully created player with resp: %v", player.String())

	return &player, nil
}

func (s *GrpcServer) GetGame(ctx context.Context, in *service.GetGameReq) (*service.Game, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	row, err := s.queries.GetGame(ctx, in.Id)
	if err != nil {
		log.Fatalf("failed to get game: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get game for id: %d", in.Id)
	}

	game := MapGetGame(row)
	log.Printf("successfully fetched game: %v", game.String())

	return game, nil
}

func (s *GrpcServer) GetGames(ctx context.Context, in *service.GetGamesReq) (*service.Games, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	// prepare arguments for fetching the games from the database
	var xPlayerParam pgtype.Int8
	var oPlayerParam pgtype.Int8

	if in.XPlayer != 0 {
		xPlayerParam.Valid = true
		xPlayerParam.Int64 = in.XPlayer
	}
	if in.OPlayer != 0 {
		oPlayerParam.Valid = true
		oPlayerParam.Int64 = in.OPlayer
	}

	params := database.GetGamesParams{
		ID:      int64(in.Page * limit),
		XPlayer: xPlayerParam,
		OPlayer: oPlayerParam,
		Limit:   int32(limit),
	}

	gameIds := make([]int64, 0)
	for i := in.Page * limit; i < (in.Page+1)*limit; i++ {
		gameIds = append(gameIds, int64(i))
	}

	// fetch games and steps from the database at the same time
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	var gameRows []database.GetGamesRow
	var stepRows []database.GameStep

	eg.Go(func() error {
		var err error
		gameRows, err = s.queries.GetGames(ctx, params)
		if err != nil {
			log.Fatalf("failed to get games: %v", err)
			return status.Errorf(codes.Internal, "failed to get games for params: %v", params)
		}
		return nil
	})
	eg.Go(func() error {
		var err error
		stepRows, err = s.queries.GetSteps(ctx, gameIds)
		if err != nil {
			log.Fatalf("failed to get steps: %v", err)
			return status.Errorf(codes.Internal, "failed to get steps for ids: %v", gameIds)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatalf("failed to retrieve games and steps for params=%v and ids=%v with err %v", params, gameIds, err)
		return nil, err
	}

	// construct a response and a log for each game
	games := MapGetGames(gameRows, stepRows)

	var sb strings.Builder
	for i, game := range games {
		sb.WriteString(game.String())
		if i < len(games) {
			sb.WriteRune(',')
			sb.WriteRune(' ')
		}
	}
	log.Printf("retrieved games with for page=%d, firstPlayer=%d, secondPlayer=%d with value: [%s]", in.Page, in.XPlayer, in.OPlayer, sb.String())

	return &service.Games{Games: games}, nil
}

func (s *GrpcServer) ListenSteps(in *service.GetGameReq, stream grpc.ServerStreamingServer[service.Step]) error {
	ctx := stream.Context()

	ticker := time.NewTicker(time.Second * 2)
	for t := range ticker.C {
		row, err := s.queries.GetLastStep(ctx, in.Id)
		if err != nil {
			log.Fatalf("failed to get last step: %v", err)
			return status.Errorf(codes.Internal, "failed to listen steps for game id: %d", in.Id)
		}

		step := MapStep(row)

		board := tictactoe.Board{}
		copy(board[:], step.Board)
		log.Fatalf("recieved step on time: %s, with value: %v, board: %v", t, step, tictactoe.BoardToString(board))

		stream.Send(step)

		if step.Result != tictactoe.Playing {
			ticker.Stop()
			return nil
		}
	}
	return nil
}

func (s *GrpcServer) Login(ctx context.Context, in *service.CredentialsReq) (*service.AuthToken, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	// retrieve if the user credentials are valid
	verify := database.VerifyPlayerParams{
		Username: in.Username,
		Passwd:   in.Password,
	}

	rows, verifyErr := s.queries.VerifyPlayer(ctx, verify)
	if verifyErr != nil {
		log.Fatalf("failed to verify player: %v", verifyErr)
		return nil, status.Errorf(codes.Internal, "failed to verify player for params: %s", verify)
	}

	if len(rows) == 0 {
		log.Printf("failed to authorize the user for username: %v", verify.Username)
		return nil, status.Errorf(codes.PermissionDenied, "authorization credentials are invalid or missing")
	} else if len(rows) > 1 {
		log.Printf("expected VerifyPlayerRows to be less than or equal to 1, was %d", len(rows))
	}

	// insert the session from the logged in player
	verifiedPlayer := rows[0]
	token := uuid.New().String()
	session := database.InsertSessionParams{
		Token:    token,
		PlayerID: verifiedPlayer.ID,
	}

	_, sessErr := s.queries.InsertSession(ctx, session)
	if sessErr != nil {
		log.Fatalf("failed to insert session: %v", sessErr)
		return nil, status.Errorf(codes.Internal, "failed to insert session for params: %v", session)
	}

	resp := service.AuthToken{Token: token}

	return &resp, nil
}

func (s *GrpcServer) MakeMove(ctx context.Context, in *service.MakeMoveReq) (*service.Game, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	token := in.GetToken().Token

	// fetch session and game at the same time
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	var sessRow database.GetSessionRow
	var gameRow database.GetGameRow

	eg.Go(func() error {
		var err error
		sessRow, err = s.queries.GetSession(ctx, token)
		if err != nil {
			log.Fatalf("failed to get session: %v", err)
			return status.Errorf(codes.Internal, "failed to get session for params: %s", token)
		}
		return nil
	})
	eg.Go(func() error {
		var err error
		gameRow, err = s.queries.GetGame(ctx, in.GameId)
		if err != nil {
			log.Fatalf("failed to get game: %v", err)
			return status.Errorf(codes.Internal, "failed to get game for id: %d", in.GameId)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatalf("failed to retrieve session and game for session token=%s and gameId=%d, with err %v", token, in.GameId, err)
		return nil, err
	}

	// perform validations for the input based on game state
	if gameRow.Result != tictactoe.Playing {
		return nil, status.Errorf(codes.InvalidArgument, "cannot make move on game: %d, game is not in play", in.GameId)
	}
	if gameRow.XTurn.Bool && gameRow.XPlayer != sessRow.ID {
		return nil, status.Errorf(codes.InvalidArgument, "cannot make move on game: %d, it isn't player's turn, expected: %d, received: %d", in.GameId, gameRow.XPlayer, sessRow.ID)
	}
	if !gameRow.XTurn.Bool && (gameRow.OPlayer.Int64 != sessRow.ID || !gameRow.OPlayer.Valid) {
		return nil, status.Errorf(codes.InvalidArgument, "cannot make move on game: %d, it isn't player's turn, expected: %d, received: %d", in.GameId, gameRow.OPlayer.Int64, sessRow.ID)
	}

	// make state mutations on the tic tac toe board
	var tileValue uint8
	if gameRow.XTurn.Bool {
		tileValue = tictactoe.X
	} else {
		tileValue = tictactoe.O
	}

	board := tictactoe.Board{}
	copy(board[:], gameRow.BoardState)

	board, turn, err := tictactoe.MoveBoard(board, gameRow.XTurn.Bool, in.Row, in.Col, tileValue)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot make move on game: %d, %s", in.GameId, err.Error())
	}
	log.Printf("made move on game: %d, board: %v", in.GameId, tictactoe.BoardToString(board))

	result := tictactoe.GetResult(board)

	boardState := make([]byte, 9)
	copy(boardState, board[:])

	// begin transaction to persist game state and game step to the database
	gameParams := database.UpdateGameParams{
		ID:         gameRow.ID,
		BoardState: boardState,
		XTurn:      pgtype.Bool{Bool: turn, Valid: true},
		UpdatedOn:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		Result:     result,
	}
	stepParams := database.InsertStepParams{
		GameID:  gameRow.ID,
		Ord:     0,
		MoveRow: int32(in.Row),
		MoveCol: int32(in.Col),
		Board:   gameRow.BoardState,
		XTurn:   turn,
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to open UpdateGame and InsertStep transaction")
	}

	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	_, err = qtx.UpdateGame(ctx, gameParams)
	if err != nil {
		log.Fatalf("failed to update game: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update game for id: %d and params: %v", in.GameId, gameParams)
	}
	_, err = qtx.InsertStep(ctx, stepParams)
	if err != nil {
		log.Fatalf("failed to update game: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update game for id: %d and params: %v", in.GameId, gameParams)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit UpdateGame and InsertStep transaction")
	}

	// contruct the game response for client
	game := MapGetGame(gameRow)

	game.BoardState = gameParams.BoardState
	game.XTurn = gameParams.XTurn.Bool
	game.UpdatedOn = &timestamppb.Timestamp{Seconds: int64(gameParams.UpdatedOn.Time.Second())}
	game.Result = gameParams.Result

	log.Fatalf("successfully made move on game: %v", game.String())

	return game, nil
}

func (s *GrpcServer) WhoAmI(ctx context.Context, in *service.AuthToken) (*service.Player, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	token := string(in.GetToken())

	log.Printf("called WhoAmI with token: %v", token)

	sessRow, err := s.queries.GetSession(ctx, token)
	if err != nil {
		log.Fatalf("failed to get session: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get session for token: %s", token)
	}

	player := service.Player{
		Id:       sessRow.ID,
		Username: sessRow.Username,
	}

	log.Printf("successfully retrieved player %v", player.String())

	return &player, nil
}
