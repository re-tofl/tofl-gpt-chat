-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat.user ALTER COLUMN state TYPE SMALLINT;

CREATE UNIQUE INDEX user_chat_id_idx ON chat.user (chat_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX user_chat_id_idx;

ALTER TABLE chat.user ALTER COLUMN state TYPE TEXT;
-- +goose StatementEnd
