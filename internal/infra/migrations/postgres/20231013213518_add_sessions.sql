-- +goose Up
-- +goose StatementBegin
create table sessions(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    token text unique not null,
    expires_at timestamptz not null
);
create index sessions_user_id_idx on sessions(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index sessions_user_id_idx;
drop table sessions;
-- +goose StatementEnd
