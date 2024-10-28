create table if not exists "user_invitation"(
    token bytea primary key,
    user_id bigint not null,
    expired timestamp(0) with time zone not null,

    constraint fk_user foreign key (user_id) references "user"(id) on delete cascade
);
