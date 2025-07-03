-- name: GetGame :one
SELECT
    g.id,
    g.x_player,
    g.o_player,
    g.board_state,
    g.x_turn,
    g.updated_on,
    g.started_on,
    g.result,
    a1.username as x_player_name,
    a2.username as o_player_name
FROM games g
LEFT JOIN player_accounts a1 ON a1.id = g.x_player
LEFT JOIN player_accounts a2 ON a2.id = g.o_player
WHERE g.id = $1;

-- name: GetGames :many
SELECT
    g.id,
    g.x_player,
    g.o_player,
    g.board_state,
    g.x_turn,
    g.updated_on,
    g.started_on,
    g.result,
    a1.username as x_player_name,
    a2.username as o_player_name
FROM games g
LEFT JOIN player_accounts a1 ON a1.id = g.x_player
LEFT JOIN player_accounts a2 ON a2.id = g.o_player
WHERE g.id > sqlc.arg('id')
    AND (g.x_player = sqlc.narg('xPlayer') OR sqlc.narg('xPlayer') IS NULL)
    AND (g.o_player = sqlc.narg('oPlayer') OR sqlc.narg('oPlayer') IS NULL)
ORDER BY g.id ASC LIMIT sqlc.arg('limit');

-- name: InsertGame :one
INSERT INTO games (x_player, o_player, board_state, x_turn, updated_on, started_on)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: UpdateGame :execresult
UPDATE games 
SET board_state = $1, x_turn = $2, updated_on = $3, result = $4
WHERE id = $5;

-- name: GetGamesSteps :many
SELECT * FROM game_steps 
WHERE game_id = ANY (sqlc.arg('gameIds')::BIGINT[])
ORDER BY game_id, ord;

-- name: GetGameSteps :many
SELECT * FROM game_steps
WHERE game_id = $1
ORDER BY ord;

-- name: InsertStep :execresult
INSERT INTO game_steps (game_id, move_row, move_col, board, x_turn, result, ord)
VALUES ($1, $2, $3, $4, $5, $6,
        COALESCE((SELECT ord FROM game_steps WHERE game_id = $1 ORDER BY ord DESC LIMIT 1), -1) + 1);

-- name: GetLastStep :one
SELECT * FROM game_steps
WHERE game_id = $1
ORDER BY ord DESC LIMIT 1;

-- name: GetPlayers :many
SELECT id, username, (SELECT COUNT(*) FROM player_sessions s WHERE s.player_id = a.id) as cnt
FROM player_accounts a
WHERE id > $1
ORDER BY id ASC LIMIT $2;

-- name: InsertPlayer :one
INSERT INTO player_accounts (username, passwd, salt)
VALUES ($1, $2, $3)
RETURNING id, username;

-- name: InsertSession :execresult
INSERT INTO player_sessions (token, player_id)
VALUES ($1, $2);

-- name: GetSession :one
SELECT a.id, a.username FROM player_sessions s
INNER JOIN player_accounts a ON a.id = s.player_id
WHERE token = $1;

-- name: GetPlayer :one
SELECT id, username FROM player_accounts WHERE id = $1;

-- name: GetAccountByName :one
SELECT id, username, passwd FROM player_accounts WHERE UPPER(username) = UPPER($1);