begin;

create schema if not exists moving;

create type moving.property_size_enum as enum (
    'unknown',
    'studio',
    '1_bedroom',
    '2_bedrooms',
    '3_bedrooms',
    '4_plus_bedrooms',
    'commercial'
);

create type moving.order_status_enum as enum (
    'unknown',
    'created',
    'rejected',
    'in_progress',
    'done'
);

create table moving.orders (
    id              serial                    primary key,
    name            varchar(100)              not null,
    email           varchar(255)              not null,
    phone           varchar(20),
    move_date       date                      not null,
    move_from       varchar(255)              not null,
    move_to         varchar(255)              not null,
    property_size   moving.property_size_enum not null,
    status          moving.order_status_enum  not null,
    additional_info varchar(500),
    created_at      timestamp,
    updated_at      timestamp
);

create index if not exists idx_moving_requests_email     on moving.orders (email);
create index if not exists idx_moving_requests_phone     on moving.orders (phone);
create index if not exists idx_moving_requests_move_date on moving.orders (move_date);

end;
