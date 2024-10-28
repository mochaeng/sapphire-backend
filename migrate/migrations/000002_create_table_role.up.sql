create table if not exists "role"(
    id serial primary key,
    name varchar(255) not null unique,
    level int not null default 0,
    description text
);

insert into role (name, level, description) values ('user', 1, 'A user can create posts, commets, and follow others + update and delete their own posts');
insert into role (name, level, description) values ('moderator', 2, 'A moderator can do everything a user can + update other users posts');
insert into role (name, level, description) values ('admin', 3, 'An admin can do everything a moderator can + delete other users posts');
