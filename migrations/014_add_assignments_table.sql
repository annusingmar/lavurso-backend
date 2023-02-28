CREATE TABLE "assignments" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "journal_id" integer NOT NULL,
    "description" text NOT NULL,
    "deadline" date NOT NULL,
    "type" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "assignments"
    ADD CONSTRAINT "assignments_relation_1" FOREIGN KEY ("journal_id") REFERENCES "journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "assignments";