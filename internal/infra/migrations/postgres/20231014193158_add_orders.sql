-- +goose Up
-- +goose StatementBegin
create type order_status as enum ('NEW', 'REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');
create table orders(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    number numeric(20, 0) check (number >= 0 and number <= 18446744073709551615) unique not null,
    status order_status not null,
    uploaded_at timestamptz not null
);
create index order_user_id_idx on orders(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index order_user_id_idx;
drop table orders;
drop type order_status;
-- +goose StatementEnd
