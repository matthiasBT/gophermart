-- +goose Up
-- +goose StatementBegin
create table session(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    token text,
    expires_at timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE session;
-- +goose StatementEnd
