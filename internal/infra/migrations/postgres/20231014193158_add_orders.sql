-- +goose Up
-- +goose StatementBegin
create type order_status as enum ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

create table orders(
    id integer primary key generated always as identity,
    user_id integer references users(id) not null,
    number numeric(20, 0) check (number >= 0 and number <= 18446744073709551615) unique not null,
    status order_status not null,
    uploaded_at timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table orders;
-- +goose StatementEnd
