package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"log"
	"time"

	"TicTacGo/db"
	"TicTacGo/pb"
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

type GrpcServer struct {
	pb.UnimplementedTicTacGoServiceServer
	queries *db.Queries
	pool    *pgxpool.Pool
}

func (s *GrpcServer) CreateGame(ctx context.Context, in *pb.CreateGameReq) (*pb.Game, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	// get the session for the provided token
	log.Printf("called CreateGame with token: %v", in.Token)

	sessRow, err := s.queries.GetSession(ctx, in.Token)
	if err != nil {
		log.Printf("failed to retrieve session: %v", err)
		return nil, status.Errorf(codes.PermissionDenied, "failed to retrieve session for token: %v", in.Token)
	}

	// insert the newly constructed game
	timeNow := time.Now()
	board, turn := tictactoe.NewGame()

	params := db.InsertGameParams{
		XPlayer:    sessRow.ID,
		BoardState: tictactoe.BoardToString(board),
		XTurn:      pgtype.Bool{Bool: turn, Valid: true},
		UpdatedOn:  pgtype.Timestamptz{Time: timeNow},
		StartedOn:  pgtype.Timestamptz{Time: timeNow},
	}

	gameId, err := s.queries.InsertGame(ctx, params)
	if err != nil {
		log.Printf("failed to insert game: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert game for params: %+v", params)
	}
	log.Printf("inserted game with params: %+v", params)

	game := pb.Game{
		Id: gameId,
		XPlayer: &pb.Player{
			Id:       sessRow.ID,
			Username: sessRow.Username,
		},
		BoardState: params.BoardState,
		XTurn:      params.XTurn.Bool,
		UpdatedOn:  &timestamppb.Timestamp{Seconds: timeNow.Unix()},
		StartedOn:  &timestamppb.Timestamp{Seconds: timeNow.Unix()},
		Steps:      []*pb.Step{},
	}

	log.Printf("successfully created game: %+v, board: %s", game.String(), tictactoe.FmtBoard(board))

	return &game, nil
}

func (s *GrpcServer) CreatePlayer(ctx context.Context, in *pb.CredentialsReq) (*pb.Player, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	params := db.InsertPlayerParams{
		Username: in.Username,
		Passwd:   in.Password,
	}

	row, err := s.queries.InsertPlayer(ctx, params)
	if err != nil {
		log.Printf("failed to insert player: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert player for params: %s", params)
	}

	player := pb.Player{
		Id:       row.ID,
		Username: row.Username,
	}

	log.Printf("successfully created player with resp: %v", player.String())

	return &player, nil
}

func (s *GrpcServer) GetGame(ctx context.Context, in *pb.GetGameReq) (*pb.Game, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	// fetch game and steps from the database at the same time
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	var gameRow db.GetGameRow
	var stepRows []db.GameStep

	eg.Go(func() error {
		var err error
		gameRow, err = s.queries.GetGame(ctx, in.Id)
		if err != nil {
			log.Printf("failed to get game: %v", err)
			return status.Errorf(codes.Internal, "failed to get games for id: %v", in.Id)
		}
		return nil
	})
	eg.Go(func() error {
		var err error
		stepRows, err = s.queries.GetGameSteps(ctx, in.Id)
		if err != nil {
			log.Printf("failed to get steps: %v", err)
			return status.Errorf(codes.Internal, "failed to get steps for id: %v", in.Id)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Printf("failed to wait for data with err: %v", err)
		return nil, err
	}

	game := MapGetGame(gameRow, stepRows)
	log.Printf("successfully fetched game: %v", game.String())

	return game, nil
}

func (s *GrpcServer) GetGames(ctx context.Context, in *pb.GetGamesReq) (*pb.Games, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	// prepare arguments for fetching the games from the database
	var xPlayerParam pgtype.Int8
	var oPlayerParam pgtype.Int8

	if in.XPlayer != nil {
		xPlayerParam.Valid = true
		xPlayerParam.Int64 = in.XPlayer.Id
	}
	if in.OPlayer != nil {
		oPlayerParam.Valid = true
		oPlayerParam.Int64 = in.OPlayer.Id
	}

	limit := in.PerPage
	params := db.GetGamesParams{
		ID:      int64((in.Page - 1) * limit),
		XPlayer: xPlayerParam,
		OPlayer: oPlayerParam,
		Limit:   limit,
	}

	gameIds := make([]int64, 0)
	for i := (in.Page - 1) * limit; i < in.Page*limit; i++ {
		gameIds = append(gameIds, int64(i))
	}

	// fetch games and steps from the database at the same time
	eg, egCtx := errgroup.WithContext(ctx)

	var gameRows []db.GetGamesRow
	var stepRows []db.GameStep

	eg.Go(func() error {
		log.Printf("fetching games for params: %+v", params)
		var err error
		gameRows, err = s.queries.GetGames(egCtx, params)
		if err != nil {
			log.Printf("failed to get games: %v", err)
			return status.Errorf(codes.Internal, "failed to get games for params: %+v", params)
		}
		return nil
	})
	eg.Go(func() error {
		var err error
		log.Printf("fetching game steps for for ids: %+v", gameIds)
		stepRows, err = s.queries.GetGamesSteps(egCtx, gameIds)
		if err != nil {
			log.Printf("failed to get steps: %v", err)
			return status.Errorf(codes.Internal, "failed to get steps for ids: %+v", gameIds)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Printf("failed to wait for data with err: %v", err)
		return nil, err
	}

	games := MapGetGames(gameRows, stepRows)
	log.Printf("retrieved games with for page=%d, firstPlayer=%v, secondPlayer=%v with value: [%s]", in.Page, in.XPlayer, in.OPlayer, GamesString(games))

	return &pb.Games{Games: games}, nil
}

func (s *GrpcServer) ListenSteps(in *pb.GetGameReq, stream grpc.ServerStreamingServer[pb.Step]) error {
	ctx := stream.Context()

	ticker := time.NewTicker(time.Second * 2)
	for t := range ticker.C {
		row, err := s.queries.GetLastStep(ctx, in.Id)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			log.Printf("failed to get last step: %v", err)
			return status.Errorf(codes.Internal, "failed to listen steps for game id: %d", in.Id)
		}

		board, err := tictactoe.BoardFromString(row.Board)
		if err != nil {
			log.Printf("error converting board from string: %v", err)
			return status.Error(codes.Internal, "error converting board from string")
		}

		step := MapStep(row)
		log.Printf("recieved step at time: %s, with value: %s, board: %v", t, step.String(), tictactoe.FmtBoard(board))

		if err = stream.Send(step); err != nil {
			log.Printf("failed to send step: %v", err)
		}

		if step.Result != tictactoe.Playing {
			ticker.Stop()
			return nil
		}
	}
	return nil
}

func (s *GrpcServer) Login(ctx context.Context, in *pb.CredentialsReq) (*pb.AuthToken, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	// retrieve if the user credentials are valid
	verify := db.VerifyPlayerParams{
		Username: in.Username,
		Passwd:   in.Password,
	}

	row, verifyErr := s.queries.VerifyPlayer(ctx, verify)
	if errors.Is(verifyErr, pgx.ErrNoRows) {
		log.Printf("failed to authorize the user for username: %v", verify.Username)
		return nil, status.Errorf(codes.PermissionDenied, "authorization credentials are invalid or missing")
	}
	if verifyErr != nil {
		log.Printf("failed to verify player: %v", verifyErr)
		return nil, status.Errorf(codes.Internal, "failed to verify player for params: %+v", verify)
	}

	// insert the session from the logged in player
	token := uuid.New().String()
	session := db.InsertSessionParams{
		Token:    token,
		PlayerID: row.ID,
	}

	_, sessErr := s.queries.InsertSession(ctx, session)
	if sessErr != nil {
		log.Printf("failed to insert session: %v", sessErr)
		return nil, status.Errorf(codes.Internal, "failed to insert session for params: %+v", session)
	}

	resp := pb.AuthToken{Token: token}

	return &resp, nil
}

func (s *GrpcServer) MakeMove(ctx context.Context, in *pb.MakeMoveReq) (*pb.Game, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	sessRow, gameRow, err := s.GetGameAndSession(ctx, in.Token, in.GameId)
	if err != nil {
		return nil, err
	}

	if !gameRow.OPlayer.Valid {
		log.Printf("cannot make move on game: %d, wait for an opponent to join as 'O'", in.GameId)
		return nil, status.Errorf(codes.PermissionDenied, "cannot make move on game: %d, wait for an opponent to join as 'O'", in.GameId)
	}
	if gameRow.Result != tictactoe.Playing {
		log.Printf("cannot make move on game: %d, game is not in play", in.GameId)
		return nil, status.Errorf(codes.PermissionDenied, "cannot make move on game: %d, game is not in play", in.GameId)
	}
	if gameRow.XTurn.Bool && gameRow.XPlayer != sessRow.ID || !gameRow.XTurn.Bool && (gameRow.OPlayer.Int64 != sessRow.ID || !gameRow.OPlayer.Valid) {
		log.Printf("cannot make move on game: %d, it isn't player's turn, expected: %d, received: %d", in.GameId, gameRow.XPlayer, sessRow.ID)
		return nil, status.Errorf(codes.PermissionDenied, "cannot make move on game: %d, it isn't player's turn, expected: %d, received: %d", in.GameId, gameRow.XPlayer, sessRow.ID)
	}

	var tileValue uint8
	if gameRow.XTurn.Bool {
		tileValue = tictactoe.X
	} else {
		tileValue = tictactoe.O
	}
	board, err := tictactoe.BoardFromString(gameRow.BoardState)
	if err != nil {
		log.Printf("error converting board from string: %v", err)
		return nil, status.Error(codes.Internal, "error converting board from string")
	}
	board, turn, err := tictactoe.MoveBoard(board, gameRow.XTurn.Bool, in.Row, in.Col, tileValue)
	if err != nil {
		log.Printf("cannot make move on game: %d, %s", in.GameId, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result := tictactoe.GetResult(board)

	log.Printf("made move on game: %d, board: %v", in.GameId, tictactoe.FmtBoard(board))

	updtGameParams := db.UpdateGameParams{
		ID:         gameRow.ID,
		BoardState: tictactoe.BoardToString(board),
		XTurn:      pgtype.Bool{Bool: turn, Valid: true},
		UpdatedOn:  pgtype.Timestamptz(pgtype.Timestamp{Time: time.Now(), Valid: true}),
		Result:     result,
	}
	instStepParams := db.InsertStepParams{
		GameID:  gameRow.ID,
		MoveRow: in.Row,
		MoveCol: in.Col,
		Board:   gameRow.BoardState,
		XTurn:   turn,
		Result:  result,
	}
	err = s.UpdateGameTransaction(ctx, in.GameId, updtGameParams, instStepParams)
	if err != nil {
		return nil, err
	}

	game := MapGetGameWithUpdt(gameRow, updtGameParams)
	return game, nil
}

func (s *GrpcServer) WhoAmI(ctx context.Context, in *pb.AuthToken) (*pb.Player, error) {
	if in == nil {
		return nil, errors.New("expected input request to be provided, was nil")
	}

	token := in.GetToken()

	log.Printf("called WhoAmI with token: %v", token)

	sessRow, err := s.queries.GetSession(ctx, token)
	if err != nil {
		log.Printf("failed to get session: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get session for token: %s", token)
	}

	player := pb.Player{
		Id:       sessRow.ID,
		Username: sessRow.Username,
	}

	log.Printf("successfully retrieved player %v", player.String())

	return &player, nil
}
