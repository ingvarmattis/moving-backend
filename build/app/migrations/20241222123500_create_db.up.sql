begin;

create schema if not exists moving;

create table moving.orders (
    id serial primary key,
    name varchar(100) not null,
    email varchar(255) not null,
    phone varchar(20),
    move_date date not null,
    move_from varchar(255) not null,
    move_to varchar(255) not null,
    property_size varchar(50) not null,
    additional_info varchar(500)
);

create index if not exists idx_moving_requests_email on moving.orders (email);
create index if not exists idx_moving_requests_phone on moving.orders (phone);
create index if not exists idx_moving_requests_move_date on moving.orders (move_date);

end;
