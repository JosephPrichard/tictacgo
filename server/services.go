package server

import (
	"TicTacGo/db"
	"TicTacGo/tictactoe"
	"context"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type MoveResult struct {
	Board  tictactoe.Board
	Turn   bool
	Result int32
}

func MakeMove(gameRow db.GetGameRow, row, col int32) (MoveResult, error) {
	var tileValue uint8
	if gameRow.XTurn.Bool {
		tileValue = tictactoe.X
	} else {
		tileValue = tictactoe.O
	}
	board, err := tictactoe.ParseBoard(gameRow.BoardState)
	if err != nil {
		log.Printf("error converting board from string: %v", err)
		return MoveResult{}, status.Error(codes.Internal, "error converting board from string")
	}
	board, turn, err := tictactoe.MoveBoard(board, gameRow.XTurn.Bool, row, col, tileValue)
	if err != nil {
		log.Printf("cannot make move on gameRow: %d, %s", gameRow.ID, err.Error())
		return MoveResult{}, status.Error(codes.InvalidArgument, err.Error())
	}
	result := tictactoe.GetResult(board)

	return MoveResult{Board: board, Turn: turn, Result: result}, nil
}

func (s *GrpcServer) GetGameAndSession(ctx context.Context, token string, gameId int64) (db.GetSessionRow, db.GetGameRow, error) {
	eg, egCtx := errgroup.WithContext(ctx)

	var sessRow db.GetSessionRow
	var gameRow db.GetGameRow

	eg.Go(func() error {
		row, err := s.Queries.GetSession(egCtx, token)
		if err != nil {
			log.Printf("failed to get session: %v", err)
			return status.Errorf(codes.PermissionDenied, "failed to get session for params: %s", token)
		}
		sessRow = row
		return nil
	})
	eg.Go(func() error {
		row, err := s.Queries.GetGame(egCtx, gameId)
		if err != nil {
			log.Printf("failed to get game: %v", err)
			return status.Errorf(codes.Internal, "failed to get game for id: %d", gameId)
		}
		gameRow = row
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Printf("failed to retrieve session and game for session token=%s and gameId=%d, with err %v", token, gameId, err)
		return db.GetSessionRow{}, db.GetGameRow{}, err
	}

	return sessRow, gameRow, nil
}

func (s *GrpcServer) UpdateGameTrans(ctx context.Context, gameId int64, updtGameParams db.UpdateGameParams, instStepParams db.InsertStepParams) error {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := s.Pool.Begin(dbCtx)
	if err != nil {
		log.Printf("failed to acquire a connection: %v", err)
		return status.Errorf(codes.Internal, "an unexpected error occured")
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Printf("failed to rollback UpdateGame and InsertStep transaction: %v", err)
		}
	}(tx, dbCtx)
	qtx := s.Queries.WithTx(tx)

	_, err = qtx.UpdateGame(dbCtx, updtGameParams)
	if err != nil {
		log.Printf("failed to update game: %v", err)
		return status.Errorf(codes.Internal, "failed to update game for id: %d and params: %+v", gameId, updtGameParams)
	}
	_, err = qtx.InsertStep(dbCtx, instStepParams)
	if err != nil {
		log.Printf("failed to insert step: %v", err)
		return status.Errorf(codes.Internal, "failed to insert step for id: %d and params: %+v", gameId, instStepParams)
	}

	if err = tx.Commit(dbCtx); err != nil {
		return status.Errorf(codes.Internal, "failed to commit UpdateGame and InsertStep transaction")
	}

	log.Printf("executed UpdateGame and InsertStep transaction for game: %d", gameId)
	return nil
}
