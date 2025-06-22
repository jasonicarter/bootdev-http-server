-- +goose Up
create table chirps (
    id UUID primary key,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null,
    body text not null,
    user_id UUID not null,
    constraint fk_user
    foreign key (user_id)
    references users(id) on delete cascade
);

-- +goose Down
drop table chirps;