begin;

create table if not exists moving.reviews (
    id          serial        primary key,
    name        varchar(100)  not null,
    rate        int           not null,
    photo_url   varchar(255)  not null,
    text        varchar(5000) not null,
    link        varchar(255)  not null,
    created_at  timestamp     not null  default now(),
    updated_at  timestamp     not null  default now()
);

end;
