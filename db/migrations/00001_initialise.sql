-- +goose Up
CREATE TABLE livestreams
(
    livestream_id   bigint GENERATED ALWAYS AS IDENTITY,
    status          text        NOT NULL,
    stream_key      text        NOT NULL,
    title           text        NOT NULL,
    description     text        NOT NULL,
    scheduled_start timestamptz NOT NULL,
    scheduled_end   timestamptz NOT NULL,
    visibility      text        NOT NULL,
    PRIMARY KEY (livestream_id),
    UNIQUE (stream_key)
);

CREATE TABLE links
(
    link_id          bigint GENERATED ALWAYS AS IDENTITY,
    livestream_id    integer NOT NULL,
    integration_type text    NOT NULL,
    integration_id   text    NOT NULL,
    PRIMARY KEY (link_id),
    CONSTRAINT fk_livestream FOREIGN KEY (livestream_id) REFERENCES livestreams (livestream_id),
    UNIQUE (integration_type, integration_id)
);

CREATE TABLE rtmp_outputs
(
    rtmp_output_id bigint GENERATED ALWAYS AS IDENTITY,
    output_url     text NOT NULL,
    PRIMARY KEY (rtmp_output_id)
);

CREATE SCHEMA mcr;

CREATE TABLE mcr.channels
(
    channel_id          bigint GENERATED ALWAYS AS IDENTITY,
    status              text    NOT NULL,
    title               text    NOT NULL,
    url_name            text    NOT NULL UNIQUE,
    res_width           integer NOT NULL,
    res_height          integer NOT NULL,
    mixer_id            integer NOT NULL,
    program_input_id    integer NOT NULL,
    continuity_input_id integer NOT NULL,
    program_output_id   integer NOT NULL,
    PRIMARY KEY (channel_id)
);

CREATE TABLE mcr.playouts
(
    playout_id      bigint GENERATED ALWAYS AS IDENTITY,
    channel_id      bigint      NOT NULL,
    brave_input_id  integer     NOT NULL,
    source_type     text        NOT NULL,
    source_uri      text        NOT NULL,
    status          text        NOT NULL,
    title           text        NOT NULL,
    description     text        NOT NULL,
    scheduled_start timestamptz NOT NULL,
    scheduled_end   timestamptz NOT NULL,
    visibility      text        NOT NULL,
    PRIMARY KEY (playout_id),
    CONSTRAINT fk_channel FOREIGN KEY (channel_id) REFERENCES mcr.channels (channel_id)
);

CREATE SCHEMA auth;

CREATE TABLE auth.tokens
(
    token_id bigint GENERATED ALWAYS AS IDENTITY,
    value    text NOT NULL,
    PRIMARY KEY (token_id)
);

CREATE SCHEMA youtube;

CREATE TABLE youtube.accounts
(
    account_id bigint GENERATED ALWAYS AS IDENTITY,
    token_id   integer NOT NULL,
    PRIMARY KEY (account_id),
    CONSTRAINT fk_token FOREIGN KEY (token_id) REFERENCES auth.tokens (token_id)
);

CREATE TABLE youtube.broadcasts
(
    broadcast_id    text   NOT NULL,
    account_id      bigint NOT NULL,
    ingest_address  text   NOT NULL,
    ingest_key      text   NOT NULL,
    title           text   NOT NULL,
    description     text   NOT NULL,
    scheduled_start text   NOT NULL,
    scheduled_end   text   NOT NULL,
    visibility      text   NOT NULL,
    PRIMARY KEY (broadcast_id),
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES youtube.accounts (account_id)
);

-- +goose Down
DROP SCHEMA youtube CASCADE;
DROP SCHEMA auth CASCADE;
DROP SCHEMA mcr CASCADE;
DROP TABLE links;
DROP TABLE rtmp_outputs;
DROP TABLE livestreams;
