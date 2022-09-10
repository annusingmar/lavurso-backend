create table
  "public"."sessions" (
    "id" serial primary key,
    "token_hash" BYTEA not null,
    "user_id" INTEGER not null,
    "expires" TIMESTAMPTZ not null,
    "login_ip" TEXT not null,
    "login_browser" TEXT not null,
    "logged_in" TIMESTAMPTZ not null,
    "last_seen" TIMESTAMPTZ not null
  );

ALTER TABLE
  "public"."sessions"
ADD
  CONSTRAINT "sessions_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;