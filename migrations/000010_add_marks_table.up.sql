CREATE TABLE "public"."marks" (
    "id" serial PRIMARY KEY,
    "user_id" integer NOT NULL,
    "lesson_id" integer,
    "course" integer,
    "journal_id" integer NOT NULL,
    "grade_id" integer,
    "comment" text,
    "type" text NOT NULL,
    "by" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "public"."marks"
    ADD CONSTRAINT "marks_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "public"."marks"
    ADD CONSTRAINT "marks_relation_2" FOREIGN KEY ("lesson_id") REFERENCES "public"."lessons" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."marks"
    ADD CONSTRAINT "marks_relation_3" FOREIGN KEY ("grade_id") REFERENCES "public"."grades" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "public"."marks"
    ADD CONSTRAINT "marks_relation_4" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."marks"
    ADD CONSTRAINT "marks_relation_6" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE INDEX "marks_index_2" ON "public"."marks" ("user_id" ASC);

ALTER TABLE marks
    ADD CONSTRAINT lesson_mark_required_fields CHECK ( CASE WHEN TYPE IN ('lesson_grade', 'not_done', 'notice_good', 'notice_neutral', 'notice_bad', 'absent', 'late') THEN
        lesson_id IS NOT NULL AND course IS NOT NULL
    END);

ALTER TABLE marks
    ADD CONSTRAINT grade_required_fields CHECK ( CASE WHEN TYPE IN ('lesson_grade', 'course_grade', 'subject_grade') THEN
        grade_id IS NOT NULL
    END);

ALTER TABLE marks
    ADD CONSTRAINT course_grade_required_field CHECK ( CASE WHEN TYPE = 'course_grade' THEN
        course IS NOT NULL
    END);

