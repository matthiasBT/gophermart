-- +goose Up
-- +goose StatementBegin
create table users(
    id integer primary key generated always as identity,
    login text unique not null,
    password_hash bytea not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
