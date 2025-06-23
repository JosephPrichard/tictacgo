CREATE TABLE player_accounts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT NOT NULL,
    passwd TEXT NOT NULL,
    salt TEXT NOT NULL
);

CREATE TABLE player_sessions (
    token TEXT NOT NULL,
    player_id BIGINT NOT NULL REFERENCES player_accounts(id),
    PRIMARY KEY(token)
);

CREATE TABLE games (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    x_player BIGINT NOT NULL REFERENCES player_accounts(id),
    o_player BIGINT REFERENCES player_accounts(id),
    board_state BYTEA NOT NULL,
    x_turn BOOLEAN,
    updated_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    result INTEGER DEFAULT 0 NOT NULL
);

CREATE TABLE game_steps (
    game_id BIGINT NOT NULL REFERENCES games(id),
    ord INTEGER NOT NULL,
    move_row INTEGER NOT NULL,
    move_col INTEGER NOT NULL,
    board BYTEA NOT NULL,
    x_turn BOOLEAN NOT NULL,
    result INTEGER NOT NULL,
    made_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(game_id, ord)
);

CREATE UNIQUE INDEX player_accounts_username ON player_accounts(username);