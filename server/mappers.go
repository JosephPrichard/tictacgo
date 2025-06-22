package main

import (
	"TicTacGo/db"
	"TicTacGo/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapGetGameDefault(row db.GetGameRow) *service.Game {
	var secondPlayer *service.Player
	if row.OPlayer.Valid {
		secondPlayer = &service.Player{
			Id:       row.OPlayer.Int64,
			Username: row.OPlayerName.String,
		}
	}

	return &service.Game{
		Id: row.ID,
		XPlayer: &service.Player{
			Id:       row.XPlayer,
			Username: row.XPlayerName.String,
		},
		OPlayer:    secondPlayer,
		BoardState: row.BoardState,
		XTurn:      row.XTurn.Bool,
		UpdatedOn:  &timestamppb.Timestamp{Seconds: int64(row.UpdatedOn.Time.Second())},
		StartedOn:  &timestamppb.Timestamp{Seconds: int64(row.StartedOn.Time.Second())},
		Steps:      []*service.Step{},
	}
}

func MapGetGameWithUpdt(row db.GetGameRow, updt db.UpdateGameParams) *service.Game {
	var secondPlayer *service.Player
	if row.OPlayer.Valid {
		secondPlayer = &service.Player{
			Id:       row.OPlayer.Int64,
			Username: row.OPlayerName.String,
		}
	}

	return &service.Game{
		Id: row.ID,
		XPlayer: &service.Player{
			Id:       row.XPlayer,
			Username: row.XPlayerName.String,
		},
		OPlayer:    secondPlayer,
		BoardState: updt.BoardState,
		XTurn:      updt.XTurn.Bool,
		UpdatedOn:  &timestamppb.Timestamp{Seconds: int64(updt.UpdatedOn.Time.Second())},
		StartedOn:  &timestamppb.Timestamp{Seconds: int64(row.StartedOn.Time.Second())},
		Result:     updt.Result,
		Steps:      []*service.Step{},
	}
}

func MapGetGames(gameRows []db.GetGamesRow, stepRows []db.GameStep) []*service.Game {
	stepsMap := make(map[int64][]*service.Step)
	var games []*service.Game
	for _, stepRow := range stepRows {
		steps, ok := stepsMap[stepRow.GameID]
		if !ok {
			steps = []*service.Step{}
		}
		steps = append(steps, MapStep(stepRow))
		stepsMap[stepRow.GameID] = steps
	}
	for _, row := range gameRows {
		var oPlayer *service.Player
		if row.OPlayer.Valid {
			oPlayer = &service.Player{
				Id:       row.OPlayer.Int64,
				Username: row.OPlayerName.String,
			}
		}
		steps, ok := stepsMap[row.ID]
		if !ok {
			steps = []*service.Step{}
		}
		game := service.Game{
			Id: row.ID,
			XPlayer: &service.Player{
				Id:       row.XPlayer,
				Username: row.XPlayerName.String,
			},
			OPlayer:    oPlayer,
			BoardState: row.BoardState,
			XTurn:      row.XTurn.Bool,
			UpdatedOn:  &timestamppb.Timestamp{Seconds: int64(row.UpdatedOn.Time.Second())},
			StartedOn:  &timestamppb.Timestamp{Seconds: int64(row.StartedOn.Time.Second())},
			Steps:      steps,
		}
		games = append(games, &game)
	}
	return games
}

func MapStep(stepRow db.GameStep) *service.Step {
	return &service.Step{
		GameId:  stepRow.GameID,
		Ord:     stepRow.Ord,
		MoveRow: stepRow.MoveRow,
		MoveCol: stepRow.MoveCol,
		Board:   stepRow.Board,
		XTurn:   stepRow.XTurn,
		Result:  stepRow.Result,
	}
}
