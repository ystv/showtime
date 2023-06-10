-- +goose Up
-- +goose StatementBegin
CREATE TABLE mixers (
    mixer_id bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
    address TEXT NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    type TEXT NOT NULL,
    CONSTRAINT type_chk CHECK (type IN ('brave', 'obs')),
    PRIMARY KEY (mixer_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE mixers;
-- +goose StatementEnd
