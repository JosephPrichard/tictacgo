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
    board_state TEXT NOT NULL,
    x_turn BOOLEAN,
    updated_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    started_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    result INTEGER DEFAULT 0 NOT NULL
);

CREATE TABLE game_steps (
    game_id BIGINT NOT NULL REFERENCES games(id),
    ord INTEGER NOT NULL,
    move_row INTEGER NOT NULL,
    move_col INTEGER NOT NULL,
    board TEXT NOT NULL,
    x_turn BOOLEAN NOT NULL,
    result INTEGER NOT NULL,
    made_on TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY(game_id, ord)
);

CREATE INDEX player_sessions_id ON player_sessions(player_id);
CREATE UNIQUE INDEX player_accounts_names ON player_accounts(UPPER(username));