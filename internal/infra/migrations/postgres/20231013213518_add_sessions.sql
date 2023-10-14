-- +goose Up
-- +goose StatementBegin
create table sessions(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    token text unique not null,
    expires_at timestamptz not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table sessions;
-- +goose StatementEnd
