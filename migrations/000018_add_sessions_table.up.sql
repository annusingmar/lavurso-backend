CREATE TABLE "public"."sessions" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "token" bytea NOT NULL,
    "user_id" integer NOT NULL,
    "expires" timestamptz NOT NULL,
    "login_ip" text NOT NULL,
    "login_browser" text NOT NULL,
    "logged_in" timestamptz NOT NULL,
    "last_seen" timestamptz NOT NULL
);

ALTER TABLE "public"."sessions"
    ADD CONSTRAINT "sessions_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

