-- seed file for unit tests into the postgres store

-- users
insert into "user" (first_name, last_name, email, username, is_active, "password", role_id) values ('momo', 'hirai', 'momo@mail.com', 'momo', true, '123', 1);
insert into "user" (first_name, last_name, email, username, is_active, "password", role_id) values ('son', 'chaeyoung', 'chae@mail.com', 'chaee', true, '123', 1);

-- posts
insert into post (tittle, "content", user_id) values ('chaeyoung <3', 'chaeyoung > lalisa', 1);