-- +goose Up
-- +goose StatementBegin
create table users(
    id integer primary key generated always as identity,
    login text unique,
    password_hash bytea
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
