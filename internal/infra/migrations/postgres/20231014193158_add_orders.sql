-- +goose Up
-- +goose StatementBegin
create type order_status as enum ('NEW', 'REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');
create table orders(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    number text unique not null,
    status order_status not null,
    uploaded_at timestamptz not null
);
create index user_orders_idx on orders(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index user_orders_idx;
drop table orders;
drop type order_status;
-- +goose StatementEnd
