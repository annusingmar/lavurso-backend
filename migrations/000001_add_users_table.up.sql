CREATE TABLE IF NOT EXISTS
  users (
    "id" serial primary key,
    "name" TEXT not null,
    "email" CITEXT UNIQUE not null,
    "password" bytea not null,
    "role" TEXT not null,
    "created_at" timestamp not null default NOW(),
    "active" BOOLEAN not null default true,
    "version" INTEGER not null default 1
  );

CREATE INDEX IF NOT EXISTS users_name_idx ON users USING GIN (to_tsvector('simple', name));