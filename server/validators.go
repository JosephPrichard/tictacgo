package server

import (
	"TicTacGo/db"
	"TicTacGo/pb"
	"TicTacGo/tictactoe"
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

const MinUsernameLen = 5
const MaxUsernameLen = 20
const MinPasswordLen = 5
const MaxPasswordLen = 100

func ValidateRegistration(in *pb.CredentialsReq) error {
	var violations []*errdetails.BadRequest_FieldViolation
	if len(in.Username) < MinUsernameLen || len(in.Username) > MaxUsernameLen {
		violation := &errdetails.BadRequest_FieldViolation{
			Field:  "username",
			Reason: fmt.Sprintf("username must be between %d and %d chars", MinUsernameLen, MaxUsernameLen),
		}
		violations = append(violations, violation)
	}
	if len(in.Password) < MinPasswordLen || len(in.Password) > MaxPasswordLen {
		violation := &errdetails.BadRequest_FieldViolation{
			Field:  "password",
			Reason: fmt.Sprintf("password must be between %d and %d chars", MinUsernameLen, MinUsernameLen),
		}
		violations = append(violations, violation)
	}

	if len(violations) == 0 {
		return nil
	}

	violation := &errdetails.BadRequest{FieldViolations: violations}
	st, err := status.New(codes.InvalidArgument, "registration credentials are invalid").WithDetails(violation)
	if err != nil {
		return err
	}
	return st.Err()
}

func ValidateMakeMove(gameRow db.GetGameRow, moverID int64) error {
	var violations []*errdetails.PreconditionFailure_Violation
	if !gameRow.OPlayer.Valid {
		violation := &errdetails.PreconditionFailure_Violation{
			Type:        "validation",
			Subject:     "state",
			Description: fmt.Sprintf("cannot make move on game: %d, wait for an opponent to join as 'O'", gameRow.ID),
		}
		violations = append(violations, violation)
		log.Printf("cannot make move on game: %d, wait for an opponent to join as 'O'", gameRow.ID)
	}
	if gameRow.Result != tictactoe.Playing {
		violation := &errdetails.PreconditionFailure_Violation{
			Type:        "validation",
			Subject:     "state",
			Description: fmt.Sprintf("cannot make move on game: %d, game is not in play", gameRow.ID),
		}
		violations = append(violations, violation)
		log.Printf("cannot make move on game: %d, game is not in play", gameRow.ID)
	}
	if gameRow.XTurn.Bool && gameRow.XPlayer != moverID || !gameRow.XTurn.Bool && (gameRow.OPlayer.Int64 != moverID || !gameRow.OPlayer.Valid) {
		violation := &errdetails.PreconditionFailure_Violation{
			Type:        "validation",
			Subject:     "turn",
			Description: fmt.Sprintf("cannot make move on game: %d, it isn't player's turn, expected: %d, received: %d", gameRow.ID, gameRow.XPlayer, moverID),
		}
		violations = append(violations, violation)
		log.Printf("cannot make move on game: %d, it isn't player's turn, expected: %d, received: %d", gameRow.ID, gameRow.XPlayer, moverID)
	}

	if len(violations) == 0 {
		return nil
	}

	violation := &errdetails.PreconditionFailure{Violations: violations}
	st, err := status.New(codes.PermissionDenied, "registration credentials are invalid").WithDetails(violation)
	if err != nil {
		return err
	}
	return st.Err()
}
