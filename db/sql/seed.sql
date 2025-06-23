INSERT INTO player_accounts (username, passwd, salt) VALUES ('user1', 'password123', 'test');
INSERT INTO player_accounts (username, passwd, salt) VALUES ('user2', 'password123', 'test');
INSERT INTO player_accounts (username, passwd, salt) VALUES ('user3', 'password123', 'test');

INSERT INTO player_sessions (token, player_id) VALUES ('User1Token', 1);

INSERT INTO games (x_player, o_player, board_state, x_turn, result)
VALUES (1, 2, E'\\x010002000000000000', true, 0);
INSERT INTO games (x_player, o_player, board_state, x_turn, result)
VALUES (2, 1, E'\\x000000000000000000', true, 0);
INSERT INTO games (x_player, o_player, board_state, x_turn, result)
VALUES (1, 3, E'\\x000000000000000000', true, 4);

INSERT INTO game_steps (game_id, ord, move_row, move_col, board, x_turn, result)
VALUES (1, 0, 0, 0, E'\\x010000000000000000', false, 0);
INSERT INTO game_steps (game_id, ord, move_row, move_col, board, x_turn, result)
VALUES (1, 1, 0, 2, E'\\x010002000000000000', true, 0);

INSERT INTO game_steps (game_id, ord, move_row, move_col, board, x_turn, result)
VALUES (3, 0, 0, 0, E'\\x000000000000000000', true, 4);