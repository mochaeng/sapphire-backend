CREATE TABLE IF NOT EXISTS "user_profile" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    description TEXT,
    avatar_url TEXT,
    banner_url TEXT,
    location VARCHAR(100),
    user_link VARCHAR(255),
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW (),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE SET NULL
);
