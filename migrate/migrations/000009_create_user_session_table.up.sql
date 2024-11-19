CREATE TABLE if not exists "user_session" (
    id TEXT NOT NULL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    constraint fk_user_id foreign key (user_id) references "user"(id)
);
