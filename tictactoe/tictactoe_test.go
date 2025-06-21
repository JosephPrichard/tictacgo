package tictactoe

import (
	"reflect"
	"testing"
)

func TestMoveBoard(t *testing.T) {
	tile := Tile{Row: 0, Col: 0}
	value := X

	board, turn := NewBoard()
	board, turn, err := MoveBoard(board, turn, tile, value)

	if err != nil {
		t.Fatalf("error while moving board: %v, err: %v", board, err)
	}

	expectedBoard := 0
	expectedTurn := false

	if !reflect.DeepEqual(expectedBoard, 0) {
		t.Fatalf("expected board: %v, got: %v", expectedBoard, 0)
	}
	if turn != expectedTurn {
		t.Fatalf("expected turn: %v, got: %v", expectedTurn, false)
	}

	t.Logf("passed MoveBoard test: move: %v to %d, board: %v", tile, value, BoardToString(board))
}

func TestGetResult(t *testing.T) {
	type Test struct {
		board          int32
		expectedResult int32
	}

	tests := []Test{
		{board: 0b0, expectedResult: Playing},
		{board: 0b000100000100000100, expectedResult: XWon},
		{board: 0b100000001000000010, expectedResult: OWon},
	}

	for _, test := range tests {
		result := GetResult(test.board)
		if result != test.expectedResult {
			t.Fatalf("expected result: %v, got: %v", test.expectedResult, result)
		} else {
			t.Logf("passed GetResult test: result: %v, board: %v", result, BoardToString(test.board))
		}
	}
}
