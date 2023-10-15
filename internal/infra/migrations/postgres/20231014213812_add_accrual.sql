-- +goose Up
-- +goose StatementBegin
create table accrual(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    order_id integer references orders(id) not null,
    amount float not null
);
create index accrual_user_id_idx on accrual(user_id);
create index accrual_order_id_idx on accrual(order_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index accrual_user_id_idx;
drop index accrual_order_id_idx;
drop table accrual;
-- +goose StatementEnd
