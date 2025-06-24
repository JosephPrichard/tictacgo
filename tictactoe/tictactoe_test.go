package tictactoe

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestMoveBoard(t *testing.T) {
	type Test struct {
		board    Board
		tile     Tile
		value    uint8
		expBoard Board
		expErr   error
	}

	tests := []Test{
		{board: Board{E, X, E, E, X, E, E, X, E}, tile: Tile{Row: 1, Col: 2}, value: X, expBoard: Board{E, X, E, E, X, X, E, X, E}},
		{board: Board{O, E, E, E, O, E, E, E, O}, tile: Tile{Row: 2, Col: 1}, value: O, expBoard: Board{O, E, E, E, O, E, E, O, O}},
		{board: Board{X, O, X, X, O, O, O, X, O}, tile: Tile{Row: 1, Col: 0}, value: O, expErr: ErrOccupied},
	}

	for i, test := range tests {
		board, turn, err := MoveBoard(test.board, false, test.tile.Row, test.tile.Col, test.value)

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Logf("ran test for board\ninput: %v\noutput: %v\nexpected: %v", FmtBoard(test.board), FmtBoard(board), FmtBoard(test.expBoard))

			if !errors.Is(err, test.expErr) {
				t.Fatalf("error while calling MoveBoard, expected err: %v, got: %v", test.expErr, err)
			} else if err == nil {
				if !reflect.DeepEqual(test.expBoard, board) {
					t.Fatalf("expected board: %v, got: %v", test.expBoard, board)
				}
				if !turn {
					t.Fatalf("expected turn: %v, got: %v", false, turn)
				}
			}
			t.Logf("passed MoveBoard test: move: %v to %d, board: %v", test.tile, test.value, FmtBoard(board))
		})
	}
}

func TestGetResult(t *testing.T) {
	type Test struct {
		board     Board
		expResult int32
	}

	tests := []Test{
		{board: Board{}, expResult: Playing},
		{board: Board{E, X, E, E, X, E, E, X, E}, expResult: XWon},
		{board: Board{O, E, E, E, O, E, E, E, O}, expResult: OWon},
		{board: Board{X, O, X, X, O, O, O, X, O}, expResult: Draw},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) {
			t.Logf("running test for board: %v", FmtBoard(test.board))

			result := GetResult(test.board)
			if result != test.expResult {
				t.Fatalf("expected result: %v, got: %v", test.expResult, result)
			} else {
				t.Logf("passed GetResult test: result: %v, board: %v", result, FmtBoard(test.board))
			}
		})
	}
}
