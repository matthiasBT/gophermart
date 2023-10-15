-- +goose Up
-- +goose StatementBegin
create table withdrawals(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    order_id integer references orders(id) not null,
    amount float not null
);
create index withdrawals_user_id_idx on withdrawals(user_id);
create index withdrawals_order_id_idx on withdrawals(order_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index withdrawals_order_id_idx;
drop index withdrawals_user_id_idx;
drop table withdrawals;
-- +goose StatementEnd
