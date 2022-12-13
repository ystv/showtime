--FIXME: run `goose fix`

-- +goose Up
CREATE TABLE livestream_events (
   livestream_event_id BIGINT GENERATED ALWAYS AS IDENTITY,
   livestream_id BIGINT NOT NULL REFERENCES livestreams(livestream_id) ON DELETE CASCADE,
   event_type TEXT NOT NULL,
   event_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   event_data JSONB DEFAULT '{}'::jsonb,
   PRIMARY KEY (livestream_event_id),
   CHECK (event_type IN (
     'started',
     'ended',
     'linked',
     'unlinked',
     'streamReceived',
     'streamLost',
     'error'
   ))
);

-- +goose Down
DROP TABLE livestream_events;
