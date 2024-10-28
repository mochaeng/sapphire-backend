create table if not exists "follower"(
    follower_id bigint not null,
    followed_id bigint not null,
    created_at timestamp(0) with time zone not null default now(),

    primary key (follower_id, followed_id),
    constraint fk_follower_user foreign key (follower_id) references "user"(id) on delete cascade,
    constraint fk_followed_user foreign key (followed_id) references "user"(id) on delete cascade,
    constraint no_self_follow check (follower_id <> followed_id)
);
