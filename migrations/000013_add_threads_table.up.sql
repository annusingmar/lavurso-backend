CREATE TABLE IF NOT EXISTS "public"."threads" (
    "id" serial PRIMARY KEY,
    "user_id" integer NOT NULL,
    "title" text NOT NULL,
    "locked" boolean NOT NULL DEFAULT FALSE,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "public"."threads"
    ADD CONSTRAINT "threads_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

