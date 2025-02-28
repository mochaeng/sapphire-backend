create table if not exists "user"(
    id bigserial primary key,
    first_name varchar(255) not null,
    last_name varchar(255) not null,
    email citext unique not null,
    username citext unique not null,
    password bytea,
    is_active boolean not null default false,
    role_id int not null,
    created_at timestamp(0) with time zone not null default now(),

    constraint fk_role foreign key (role_id) references "role"(id)
);
