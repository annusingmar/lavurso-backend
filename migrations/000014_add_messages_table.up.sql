CREATE TABLE IF NOT EXISTS "public"."messages" (
    "id" serial PRIMARY KEY,
    "thread_id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "body" text NOT NULL,
    "type" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "public"."messages"
    ADD CONSTRAINT "messages_relation_1" FOREIGN KEY ("thread_id") REFERENCES "public"."threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."messages"
    ADD CONSTRAINT "messages_relation_2" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

