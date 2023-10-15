-- +goose Up
-- +goose StatementBegin
create table withdrawals(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    order_number numeric(20, 0),
    amount float not null,
    processed_at timestamptz not null
);
create index withdrawals_user_id_idx on withdrawals(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index withdrawals_user_id_idx;
drop table withdrawals;
-- +goose StatementEnd
