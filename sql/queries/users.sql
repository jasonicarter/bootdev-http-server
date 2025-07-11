-- name: CreateUser :one
insert into users (id, created_at, updated_at, email, hashed_password)
values (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
returning *;

-- name: ResetUsers :exec
truncate table users;

-- name: LoginUser :one
select * from users where email = $1;