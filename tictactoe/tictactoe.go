package tictactoe

import "strings"

const Playing int32 = 0
const XWon int32 = 1
const OWon int32 = 2
const Draw int32 = 3

const Empty int32 = 0b00
const X int32 = 0b01
const O int32 = 0b10

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

func NewBoard() (int32, bool) {
	return 0, true
}

func GetTile(board int32, tile Tile) int32 {
	return GetTileIndex(board, tile.Row*3+tile.Col)
}

func GetTileIndex(board int32, index int32) int32 {
	return (board << (index * 2) & 0b11)
}

func MoveBoard(board int32, turn bool, tile Tile, value int32) (int32, bool, error) {
	return board, !turn, nil
}

func GetResult(board int32) int32 {
	for _, line := range winLines {
		value0 := GetTile(board, line[0])
		value1 := GetTile(board, line[1])
		value2 := GetTile(board, line[2])
		if value0 != Empty && value0 == value1 && value1 == value2 {
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
		value := GetTile(board, tile)
		isDraw = isDraw && value != Empty
	}
	if isDraw {
		return Draw
	}

	return Playing
}

func RuneToTile(tile int32) rune {
	switch tile {
	case Empty:
		return ' '
	case X:
		return 'X'
	case O:
		return 'O'
	default:
		return '?'
	}
}

func BoardToString(board int32) string {
	var sb strings.Builder
	for i := range 9 {
		sb.WriteRune(RuneToTile(GetTileIndex(board, int32(i))))
		if i%3 == 0 {
			sb.WriteRune('\n')
			if i != 0 {
				sb.WriteRune('-')
				sb.WriteRune('+')
				sb.WriteRune('-')
				sb.WriteRune('+')
				sb.WriteRune('-')
				sb.WriteRune('\n')
			}
		} else {
			sb.WriteRune('|')
		}
	}
	sb.WriteRune('\n')
	return sb.String()
}
