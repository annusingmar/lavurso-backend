CREATE TABLE IF NOT EXISTS
  users (
    "id" serial primary key,
    "name" TEXT not null,
    "email" CITEXT UNIQUE not null,
    "phone_number" TEXT,
    "id_code" BIGINT,
    "birth_date" DATE,
    "password" bytea not null,
    "role" TEXT not null,
    "class_id" INTEGER,
    "created_at" timestamp not null default NOW(),
    "active" BOOLEAN not null default true,
    "archived" BOOLEAN not null default false
  );

CREATE INDEX IF NOT EXISTS users_name_idx ON users USING GIN (to_tsvector('simple', name));

ALTER TABLE users ADD CONSTRAINT student_class_id_not_null CHECK(CASE WHEN role = 'student' THEN class_id is NOT NULL END);