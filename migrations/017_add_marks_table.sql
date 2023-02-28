CREATE TABLE "marks" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id" integer NOT NULL,
    "lesson_id" integer,
    "course" integer,
    "journal_id" integer NOT NULL,
    "grade_id" integer,
    "comment" text,
    "type" text NOT NULL,
    "teacher_id" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "marks"
    ADD CONSTRAINT "marks_relation_1" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "marks"
    ADD CONSTRAINT "marks_relation_2" FOREIGN KEY ("lesson_id") REFERENCES "lessons" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "marks"
    ADD CONSTRAINT "marks_relation_3" FOREIGN KEY ("grade_id") REFERENCES "grades" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "marks"
    ADD CONSTRAINT "marks_relation_4" FOREIGN KEY ("journal_id") REFERENCES "journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "marks"
    ADD CONSTRAINT "marks_relation_5" FOREIGN KEY ("teacher_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE marks
    ADD CONSTRAINT lesson_mark_required_fields CHECK ( CASE WHEN type IN ('lesson_grade', 'not_done', 'notice_good', 'notice_neutral', 'notice_bad', 'absent', 'late') THEN
        lesson_id IS NOT NULL AND course IS NOT NULL
    END);

ALTER TABLE marks
    ADD CONSTRAINT grade_required_fields CHECK ( CASE WHEN type IN ('lesson_grade', 'course_grade', 'subject_grade') THEN
        grade_id IS NOT NULL
    END);

ALTER TABLE marks
    ADD CONSTRAINT course_grade_required_field CHECK ( CASE WHEN type = 'course_grade' THEN
        course IS NOT NULL
    END);

CREATE UNIQUE INDEX "only_one_per_student_per_lesson" ON marks (user_id, lesson_id, type) WHERE (type IN ('absent', 'late', 'not_done'));

---- create above / drop below ----

DROP TABLE "marks";