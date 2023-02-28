CREATE TABLE "lessons" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "journal_id" integer NOT NULL,
    "description" text NOT NULL,
    "date" date NOT NULL,
    "course" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "lessons"
    ADD CONSTRAINT "lessons_relation_1" FOREIGN KEY ("journal_id") REFERENCES "journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "lessons";