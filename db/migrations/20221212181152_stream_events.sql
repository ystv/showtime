--FIXME: run `goose fix`

-- +goose Up
CREATE TYPE livestream_event_type AS ENUM (
    'started',
    'ended',
    'linked',
    'unlinked',
    'stream_received',
    'stream_lost',
    'error'
);
CREATE TABLE livestream_events (
   livestream_event_id BIGSERIAL PRIMARY KEY,
   livestream_id integer NOT NULL REFERENCES livestreams(livestream_id),
   event_type livestream_event_type NOT NULL,
   event_time timestamptz NOT NULL DEFAULT NOW(),
   event_data jsonb DEFAULT '{}'::jsonb
);

-- +goose Down
DROP TABLE livestream_events;
DROP TYPE livestream_event_type;
