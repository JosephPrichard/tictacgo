package server

import (
	"TicTacGo/db"
	"TicTacGo/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
)

func MapGetGame(gameRow db.GetGameRow, stepRows []db.GameStep) *pb.Game {
	var steps []*pb.Step
	for _, stepRow := range stepRows {
		steps = append(steps, MapStep(stepRow))
	}

	var secondPlayer *pb.Player
	if gameRow.OPlayer.Valid {
		secondPlayer = &pb.Player{
			Id:       gameRow.OPlayer.Int64,
			Username: gameRow.OPlayerName.String,
		}
	}

	return &pb.Game{
		Id: gameRow.ID,
		XPlayer: &pb.Player{
			Id:       gameRow.XPlayer,
			Username: gameRow.XPlayerName.String,
		},
		OPlayer:    secondPlayer,
		BoardState: gameRow.BoardState,
		XTurn:      gameRow.XTurn.Bool,
		UpdatedOn:  &timestamppb.Timestamp{Seconds: gameRow.UpdatedOn.Time.Unix()},
		StartedOn:  &timestamppb.Timestamp{Seconds: gameRow.StartedOn.Time.Unix()},
		Result:     gameRow.Result,
		Steps:      steps,
	}
}

func MapGetGameWithUpdt(row db.GetGameRow, updt db.UpdateGameParams) *pb.Game {
	var oPlayer *pb.Player
	if row.OPlayer.Valid {
		oPlayer = &pb.Player{
			Id:       row.OPlayer.Int64,
			Username: row.OPlayerName.String,
		}
	}

	return &pb.Game{
		Id: row.ID,
		XPlayer: &pb.Player{
			Id:       row.XPlayer,
			Username: row.XPlayerName.String,
		},
		OPlayer:    oPlayer,
		BoardState: updt.BoardState,
		XTurn:      updt.XTurn.Bool,
		UpdatedOn:  &timestamppb.Timestamp{Seconds: updt.UpdatedOn.Time.Unix()},
		StartedOn:  &timestamppb.Timestamp{Seconds: row.StartedOn.Time.Unix()},
		Result:     updt.Result,
		Steps:      []*pb.Step{},
	}
}

func MapGetGames(gameRows []db.GetGamesRow, stepRows []db.GameStep) []*pb.Game {
	stepsMap := make(map[int64][]*pb.Step)
	var games []*pb.Game

	for _, stepRow := range stepRows {
		steps, ok := stepsMap[stepRow.GameID]
		if !ok {
			steps = []*pb.Step{}
		}
		steps = append(steps, MapStep(stepRow))
		stepsMap[stepRow.GameID] = steps
	}
	for _, row := range gameRows {
		var oPlayer *pb.Player
		if row.OPlayer.Valid {
			oPlayer = &pb.Player{
				Id:       row.OPlayer.Int64,
				Username: row.OPlayerName.String,
			}
		}
		steps, ok := stepsMap[row.ID]
		if !ok {
			steps = []*pb.Step{}
		}
		game := pb.Game{
			Id: row.ID,
			XPlayer: &pb.Player{
				Id:       row.XPlayer,
				Username: row.XPlayerName.String,
			},
			OPlayer:    oPlayer,
			BoardState: row.BoardState,
			XTurn:      row.XTurn.Bool,
			UpdatedOn:  &timestamppb.Timestamp{Seconds: row.UpdatedOn.Time.Unix()},
			StartedOn:  &timestamppb.Timestamp{Seconds: row.StartedOn.Time.Unix()},
			Result:     row.Result,
			Steps:      steps,
		}
		games = append(games, &game)
	}

	return games
}

func MapStep(stepRow db.GameStep) *pb.Step {
	return &pb.Step{
		GameId:  stepRow.GameID,
		Ord:     stepRow.Ord,
		MoveRow: stepRow.MoveRow,
		MoveCol: stepRow.MoveCol,
		Board:   stepRow.Board,
		XTurn:   stepRow.XTurn,
		Result:  stepRow.Result,
	}
}

func MapPlayers(rows []db.GetPlayersRow) []*pb.Player {
	var players []*pb.Player

	for _, row := range rows {
		player := &pb.Player{
			Id:       row.ID,
			Username: row.Username,
			Cnt:      int32(row.Cnt),
		}
		players = append(players, player)
	}

	return players
}

func GamesString(games []*pb.Game) string {
	var sb strings.Builder
	for i, game := range games {
		sb.WriteString(game.String())
		if i < len(games) {
			sb.WriteRune(',')
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}
