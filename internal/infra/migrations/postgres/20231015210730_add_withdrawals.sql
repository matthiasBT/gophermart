-- +goose Up
-- +goose StatementBegin
create table withdrawals(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    order_number text not null,
    amount float not null,
    processed_at timestamptz not null
);
create index user_withdrawals_idx on withdrawals(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index user_withdrawals_idx;
drop table withdrawals;
-- +goose StatementEnd
