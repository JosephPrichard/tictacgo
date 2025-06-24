package tictactoe

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"
)

const (
	Playing int32 = 0
	XWon    int32 = 1
	OWon    int32 = 2
	Draw    int32 = 3
	Forfeit int32 = 4
)

const (
	E uint8 = 0b00
	X uint8 = 0b01
	O uint8 = 0b10
)

var (
	topLeft     = Tile{Row: 0, Col: 0}
	top         = Tile{Row: 0, Col: 1}
	topRight    = Tile{Row: 0, Col: 2}
	left        = Tile{Row: 1, Col: 0}
	middle      = Tile{Row: 1, Col: 1}
	right       = Tile{Row: 1, Col: 2}
	bottomLeft  = Tile{Row: 2, Col: 0}
	bottom      = Tile{Row: 2, Col: 1}
	bottomRight = Tile{Row: 2, Col: 2}
	allLines    = []Tile{topLeft, top, topRight, left, middle, right, bottomLeft, bottom, bottomRight}
)

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

type Board = [9]uint8

type Tile struct {
	Row int32
	Col int32
}

func NewGame() (Board, bool) {
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

func FmtBoard(board Board) string {
	var sb strings.Builder
	sb.WriteRune('\n')
	for i := range 9 {
		tile := board[i]

		var c rune
		switch tile {
		case E:
			c = ' '
			break
		case X:
			c = 'X'
			break
		case O:
			c = 'O'
			break
		default:
			c = '?'
			break
		}
		sb.WriteRune(c)

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

func BoardToString(board Board) string {
	var sb strings.Builder
	for _, tile := range board {
		var c rune
		switch tile {
		case E:
			c = '_'
			break
		case X:
			c = 'x'
			break
		case O:
			c = 'o'
			break
		default:
			log.Printf("attempting to convert invalid board to string: %v", board)
			c = '?'
			break
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

func BoardFromString(s string) (Board, error) {
	var board Board
	for i, c := range s {
		if i >= len(board) {
			return Board{}, fmt.Errorf("attempting to parse invalid board from string: %s, length exceeds maximum: %d", s, len(board))
		}
		var tile uint8
		switch unicode.ToLower(c) {
		case '_':
			tile = E
			break
		case 'x':
			tile = X
			break
		case 'o':
			tile = O
			break
		default:
			return Board{}, fmt.Errorf("attempting to parse invalid board from string: %s, invalid symbol: %v at %d", s, c, i)
		}
		board[i] = tile
	}
	return board, nil
}
