-- name: GetGame :one
SELECT * FROM games
WHERE id = $1;

-- name: GetGames :many
SELECT * FROM games
WHERE id > $1 AND (first_player = COALESCE($2, first_player) OR second_player = COALESCE($3, second_player)) AND result != 0
ORDER BY id ASC LIMIT $4;

-- name: InsertGame :execresult
INSERT INTO 
games (first_player, second_player, creating_player, board_state, turn, updated_on, started_on) 
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateGame :execresult
UPDATE games 
SET board_state = $1, turn = $2, updated_on = $3, result = $4
WHERE id = $5;

-- name: GetSteps :many
SELECT * FROM game_steps 
WHERE game_id = ANY ($1::BIGINT[])
ORDER BY game_id, ord;

-- name: InsertStep :execresult
INSERT INTO 
game_steps (game_id, ord, move_row, move_col, board, turn) 
VALUES ($1, $2, $3, $4, $5, $6);