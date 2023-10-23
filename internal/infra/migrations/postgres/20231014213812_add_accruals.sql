-- +goose Up
-- +goose StatementBegin
create table accruals(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    order_number text not null,
    processed_at timestamptz, -- null for accruals (positive), not null for withdrawals (negative)?
    amount float not null
);
create index user_accruals_idx on accruals(user_id);
create index order_accruals_idx on accruals(order_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index user_accruals_idx;
drop index order_accruals_idx;
drop table accruals;
-- +goose StatementEnd
