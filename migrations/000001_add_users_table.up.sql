CREATE TABLE IF NOT EXISTS
  users (
    "id" serial primary key,
    "name" TEXT not null,
    "email" CITEXT UNIQUE not null,
    "password" bytea not null,
    "phone" TEXT not null,
    "address" TEXT not null,
    "birth_date" TIMESTAMP not null,
    "role" INTEGER not null,
    "created_at" timestamp not null default NOW(),
    "version" INTEGER not null default 1
  );