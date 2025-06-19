CREATE TABLE games (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    first_player BIGINT NOT NULL,
    second_player BIGINT NOT NULL,
    creating_player BIGINT NOT NULL,
    board_state BYTEA NOT NULL,
    turn BOOLEAN,
    updated_on TIMESTAMP NOT NULL,
    started_on TIMESTAMP NOT NULL,
    result INTEGER DEFAULT 0 NOT NULL
);

CREATE TABLE game_steps (
    game_id BIGINT NOT NULL,
    ord INTEGER NOT NULL,
    move_row INTEGER NOT NULL,
    move_col INTEGER NOT NULL,
    board BYTEA NOT NULL,
    turn BOOLEAN
);