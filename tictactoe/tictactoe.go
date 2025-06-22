package tictactoe

import (
	"errors"
	"strings"
)

const Playing int32 = 0
const XWon int32 = 1
const OWon int32 = 2
const Draw int32 = 3

const E uint8 = 0b00
const X uint8 = 0b01
const O uint8 = 0b10

type Board = [9]uint8

type Tile struct {
	Row int32
	Col int32
}

var topLeft = Tile{Row: 0, Col: 0}
var top = Tile{Row: 0, Col: 1}
var topRight = Tile{Row: 0, Col: 2}
var left = Tile{Row: 1, Col: 0}
var middle = Tile{Row: 1, Col: 1}
var right = Tile{Row: 1, Col: 2}
var bottomLeft = Tile{Row: 2, Col: 0}
var bottom = Tile{Row: 2, Col: 1}
var bottomRight = Tile{Row: 2, Col: 2}

var allLines = []Tile{topLeft, top, topRight, left, middle, right, bottomLeft, bottom, bottomRight}

var winLines = [][]Tile{
	{topLeft, top, topRight},          // Row 1
	{left, middle, right},             // Row 2
	{bottomLeft, bottom, bottomRight}, // Row 3
	{topLeft, left, bottomLeft},       // Column 1
	{top, middle, bottom},             // Column 2
	{topRight, right, bottomRight},    // Column 3
	{topLeft, middle, bottomRight},    // Diagonal \
	{topRight, middle, bottomLeft},    // Diagonal /
}

func NewBoard() (Board, bool) {
	return Board{}, true
}

var ErrTile = errors.New("tile is invalid")
var ErrOccupied = errors.New("value at tile is not empty")

func GetIndex(tile Tile) int32 {
	return tile.Row*3 + tile.Col
}

func MoveBoard(board Board, turn bool, row int32, col int32, value uint8) (Board, bool, error) {
	tile := Tile{Row: row, Col: col}
	if tile.Row < 0 || tile.Col < 0 || tile.Row > 2 && tile.Col > 2 {
		return board, turn, ErrTile
	}
	index := GetIndex(tile)

	t := board[index]
	if t != E {
		return board, turn, ErrOccupied
	}

	board[index] = value
	return board, !turn, nil
}

func GetResult(board Board) int32 {
	for _, line := range winLines {
		value0 := board[GetIndex(line[0])]
		value1 := board[GetIndex(line[1])]
		value2 := board[GetIndex(line[2])]
		if value0 != E && value0 == value1 && value1 == value2 {
			switch value0 {
			case X:
				return XWon
			case O:
				return OWon
			}
		}
	}

	isDraw := true
	for _, tile := range allLines {
		value := board[GetIndex(tile)]
		isDraw = isDraw && value != E
	}
	if isDraw {
		return Draw
	}

	return Playing
}

func RuneToTile(tile uint8) rune {
	switch tile {
	case E:
		return ' '
	case X:
		return 'X'
	case O:
		return 'O'
	default:
		return '?'
	}
}

func BoardToString(board Board) string {
	var sb strings.Builder
	sb.WriteRune('\n')
	for i := range 9 {
		tile := board[i]
		runeTile := RuneToTile(tile)
		sb.WriteRune(runeTile)
		if (i+1)%3 == 0 && i != 8 {
			sb.WriteRune('\n')
			sb.WriteRune('-')
			sb.WriteRune('+')
			sb.WriteRune('-')
			sb.WriteRune('+')
			sb.WriteRune('-')
			sb.WriteRune('\n')
		} else if (i+1)%3 != 0 {
			sb.WriteRune('|')
		}
	}
	sb.WriteRune('\n')
	return sb.String()
}
