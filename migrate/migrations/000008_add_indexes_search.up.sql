create extension if not exists pg_trgm;
create extension if not exists btree_gin;

create index if not exists idx_comment_content on "comment" using gin (content gin_trgm_ops);

create index if not exists idx_post_tittle on "post" using gin (tittle gin_trgm_ops);
create index if not exists idx_post_tags on "post" using gin (tags);

create index if not exists idx_user_username on "user"(username);
create index if not exists idx_post_user_id on "post"(user_id);
create index if not exists idx_comment_post_id on "comment"(post_id);
