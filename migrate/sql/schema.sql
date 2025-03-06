create extension if not exists citext;

-- roles: used for authorization in the application
create table if not exists "role"(
    id serial primary key,
    name varchar(255) not null unique,
    level int not null default 0,
    description text
);
insert into role (name, level, description) values ('user', 1, 'A user can create posts, commets, and follow others + update and delete their own posts');
insert into role (name, level, description) values ('moderator', 2, 'A moderator can do everything a user can + update other users posts');
insert into role (name, level, description) values ('admin', 3, 'An admin can do everything a moderator can + delete other users posts');

create table if not exists "user"(
    id bigserial primary key,
    first_name varchar(255) not null,
    last_name varchar(255),
    email citext unique not null,
    username citext unique not null,
    password_hash bytea,
    is_active boolean not null default false,
    role_id int not null,
    created_at timestamp(0) with time zone not null default now(),

    constraint fk_role foreign key (role_id) references "role"(id)
);

create table if not exists "user_profile" (
    id bigserial primary key,
    user_id bigint not null,
    description text,
    avatar_url text,
    banner_url text,
    location varchar(100),
    user_link varchar(255),
    created_at timestamp(0) with time zone not null default now (),
    updated_at timestamp(0) with time zone not null default now (),
    constraint fk_user foreign key (user_id) references "user" (id) on delete set null
);

create table if not exists "user_session" (
    id text not null primary key,
    user_id bigint not null,
    expires_at timestamptz not null,

    constraint fk_user_id foreign key (user_id) references "user"(id)
);

create table if not exists "oauth_account" (
	provider_id text not null,
	provider_user_id text not null,
	user_id text not null,
	created_at timestamp(0) with time zone not null default now (),

	constraint oauth_account_provider_id_provider_user_id_pk primary key(provider_id, provider_user_id),
	constraint fk_user_id foreign key (user_id) references "user"(id)
);

create table if not exists "post"(
    id bigserial primary key,
    tittle text not null,
    user_id bigint not null,
    content text not null,
    tags varchar(255) [],
    -- media_url varchar(255),
    media_urls text [],
    media_type varchar(255),  -- photo/video/gif/many
    created_at timestamp(0) with time zone not null default now(),
    updated_at timestamp(0) with time zone not null default now(),

    constraint fk_user foreign key (user_id) references "user"(id) on delete set null
);

CREATE TABLE IF NOT EXISTS "comment"(
  id bigserial PRIMARY KEY,
  post_id bigserial,
  user_id bigserial NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),

  constraint fk_post foreign key (post_id) references "post"(id) on delete set null,
  constraint fk_user foreign key (user_id) references "user"(id) on delete set null
);

create table if not exists "follow"(
    follower_id bigint not null,
    followed_id bigint not null,
    created_at timestamp(0) with time zone not null default now(),

    primary key (follower_id, followed_id),
    constraint fk_follower_user foreign key (follower_id) references "user"(id) on delete cascade,
    constraint fk_followed_user foreign key (followed_id) references "user"(id) on delete cascade,
    constraint no_self_follow check (follower_id <> followed_id)
);

create table if not exists "user_invitation"(
    token bytea primary key,
    user_id bigint not null,
    expired timestamp(0) with time zone not null,

    constraint fk_user foreign key (user_id) references "user"(id) on delete cascade
);

-- indexes: used to increase search speed
create extension if not exists pg_trgm;
create extension if not exists btree_gin;
create index if not exists idx_comment_content on "comment" using gin (content gin_trgm_ops);
create index if not exists idx_post_tittle on "post" using gin (tittle gin_trgm_ops);
create index if not exists idx_post_tags on "post" using gin (tags);
create index if not exists idx_user_username on "user"(username);

-- create index if not exists idx_post_user_id on "post"(user_id);

create index if not exists idx_post_user_created on "post" (user_id, created_at desc);
create index if not exists idx_post_user_id on "post" (user_id, id desc);

create index if not exists idx_comment_post_id on "comment"(post_id);
