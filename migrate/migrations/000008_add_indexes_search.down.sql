drop extension if exists pg_trgm;
drop extension if exists btree_gin;

drop index if exists idx_comment_content;
drop index if exists idx_post_tittle;
drop index if exists idx_post_tags;
drop index if exists idx_user_username;
drop index if exists idx_post_user_id;
drop index if exists idx_comment_post_id;
