CREATE TABLE "students_journals" (
    "student_id" integer NOT NULL,
    "journal_id" integer NOT NULL
);

ALTER TABLE "students_journals"
    ADD CONSTRAINT "students_journals_pkey" PRIMARY KEY ("student_id", "journal_id");

ALTER TABLE "students_journals"
    ADD CONSTRAINT "students_journals_relation_1" FOREIGN KEY ("student_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "students_journals"
    ADD CONSTRAINT "students_journals_relation_2" FOREIGN KEY ("journal_id") REFERENCES "journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "students_journals";