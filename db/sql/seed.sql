INSERT INTO player_accounts (username, passwd, salt) VALUES ('user1', 'password123', 'test');
INSERT INTO player_accounts (username, passwd, salt) VALUES ('user2', 'password123', 'test');
INSERT INTO player_accounts (username, passwd, salt) VALUES ('user3', 'password123', 'test');

INSERT INTO player_sessions (token, player_id) VALUES ('User1Token', 1);
INSERT INTO player_sessions (token, player_id) VALUES ('User3Token', 3);

INSERT INTO games (x_player, o_player, board_state, x_turn, result)
VALUES (1, 2, 'x_o______', true, 0);
INSERT INTO games (x_player, o_player, board_state, x_turn, result)
VALUES (2, 1, '_________', true, 0);
INSERT INTO games (x_player, o_player, board_state, x_turn, result)
VALUES (1, 3, '_________', true, 4);
INSERT INTO games (x_player, board_state, x_turn, result)
VALUES (1, '_________', true, 0);

INSERT INTO game_steps (game_id, ord, move_row, move_col, board, x_turn, result)
VALUES (1, 0, 0, 0, 'x________', false, 0);
INSERT INTO game_steps (game_id, ord, move_row, move_col, board, x_turn, result)
VALUES (1, 1, 0, 2, 'x_o______', true, 0);

INSERT INTO game_steps (game_id, ord, move_row, move_col, board, x_turn, result)
VALUES (3, 0, 0, 0, '_________', true, 4);