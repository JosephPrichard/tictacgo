package tictactoe

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestMoveBoard(t *testing.T) {
	type Test struct {
		board         Board
		tile          Tile
		value         uint8
		expectedBoard Board
		expectedErr   error
	}

	tests := []Test{
		{board: Board{E, X, E, E, X, E, E, X, E}, tile: Tile{Row: 1, Col: 2}, value: X, expectedBoard: Board{E, X, E, E, X, X, E, X, E}},
		{board: Board{O, E, E, E, O, E, E, E, O}, tile: Tile{Row: 2, Col: 1}, value: O, expectedBoard: Board{O, E, E, E, O, E, E, O, O}},
		{board: Board{X, O, X, X, O, O, O, X, O}, tile: Tile{Row: 1, Col: 0}, value: O, expectedErr: ErrOccupied},
	}

	for i, test := range tests {
		board, turn, err := MoveBoard(test.board, false, test.tile.Row, test.tile.Col, test.value)

		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			t.Logf("ran test for board\ninput: %v\noutput: %v\nexpected: %v", BoardToString(test.board), BoardToString(board), BoardToString(test.expectedBoard))

			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("error while calling MoveBoard, expected err: %v, got: %v", test.expectedErr, err)
			} else if err == nil {
				if !reflect.DeepEqual(test.expectedBoard, board) {
					t.Fatalf("expected board: %v, got: %v", test.expectedBoard, board)
				}
				if !turn {
					t.Fatalf("expected turn: %v, got: %v", false, turn)
				}
			}
			t.Logf("passed MoveBoard test: move: %v to %d, board: %v", test.tile, test.value, BoardToString(board))
		})
	}
}

func TestGetResult(t *testing.T) {
	type Test struct {
		board          Board
		expectedResult int32
	}

	tests := []Test{
		{board: Board{}, expectedResult: Playing},
		{board: Board{E, X, E, E, X, E, E, X, E}, expectedResult: XWon},
		{board: Board{O, E, E, E, O, E, E, E, O}, expectedResult: OWon},
		{board: Board{X, O, X, X, O, O, O, X, O}, expectedResult: Draw},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			t.Logf("running test for board: %v", BoardToString(test.board))

			result := GetResult(test.board)
			if result != test.expectedResult {
				t.Fatalf("expected result: %v, got: %v", test.expectedResult, result)
			} else {
				t.Logf("passed GetResult test: result: %v, board: %v", result, BoardToString(test.board))
			}
		})
	}
}
