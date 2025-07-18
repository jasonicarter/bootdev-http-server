-- +goose Up
create table users (
    id UUID primary key,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null,
    email text not null unique
);

-- +goose Down
drop table users;