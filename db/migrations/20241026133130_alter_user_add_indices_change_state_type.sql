-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX user_chat_id_idx ON chat.user (chat_id);
ALTER TABLE chat.user DROP COLUMN state;
ALTER TABLE chat.user ADD COLUMN state SMALLINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chat.user ADD COLUMN state TEXT;
ALTER TABLE chat.user DROP COLUMN state;
DROP INDEX user_chat_id_idx;
-- +goose StatementEnd
