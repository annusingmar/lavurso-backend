CREATE TABLE "teachers_journals" (
    "teacher_id" integer NOT NULL,
    "journal_id" integer NOT NULL
);

ALTER TABLE "teachers_journals"
    ADD CONSTRAINT "teachers_journals_pkey" PRIMARY KEY ("teacher_id", "journal_id");

ALTER TABLE "teachers_journals"
    ADD CONSTRAINT "teachers_journals_relation_1" FOREIGN KEY ("teacher_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "teachers_journals"
    ADD CONSTRAINT "teachers_journals_relation_2" FOREIGN KEY ("journal_id") REFERENCES "journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "teachers_journals";