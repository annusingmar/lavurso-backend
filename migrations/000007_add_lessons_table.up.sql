create table if not exists
  "public"."lessons" (
    "id" serial primary key,
    "journal_id" INTEGER not null,
    "description" TEXT not null,
    "date" DATE not null,
    "course" INTEGER not null,
    "created_at" TIMESTAMPTZ not null default NOW(),
    "updated_at" TIMESTAMPTZ not null default NOW()
  );

ALTER TABLE
  "public"."lessons"
ADD
  CONSTRAINT "lessons_relation_1" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;