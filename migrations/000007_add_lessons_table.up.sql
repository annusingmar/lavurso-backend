CREATE TABLE IF NOT EXISTS "public"."lessons" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "journal_id" integer NOT NULL,
    "description" text NOT NULL,
    "date" date NOT NULL,
    "course" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "public"."lessons"
    ADD CONSTRAINT "lessons_relation_1" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

