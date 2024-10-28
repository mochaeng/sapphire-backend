create table if not exists "post"(
    id bigserial primary key,
    tittle text not null,
    user_id bigint not null,
    content text not null,
    tags varchar(255) [],
    media_url varchar(255),
    created_at timestamp(0) with time zone not null default now(),
    updated_at timestamp(0) with time zone not null default now(),

    constraint fk_user foreign key (user_id) references "user"(id) on delete set null
);
