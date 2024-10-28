CREATE TABLE IF NOT EXISTS "comment"(
  id bigserial PRIMARY KEY,
  post_id bigserial,
  user_id bigserial NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),

  constraint fk_post foreign key (post_id) references "post"(id) on delete set null,
  constraint fk_user foreign key (user_id) references "user"(id) on delete set null
);
