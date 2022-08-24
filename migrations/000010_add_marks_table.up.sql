create table
  "public"."marks" (
    "id" serial primary key,
    "user_id" INTEGER not null,
    "lesson_id" INTEGER,
    "course" INTEGER,
    "journal_id" INTEGER not null,
    "grade_id" INTEGER,
    "comment" TEXT,
    "type" TEXT not null,
    "by" INTEGER not null,
    "created_at" TIMESTAMP not null default NOW(),
    "updated_at" TIMESTAMP not null default NOW()
  );

ALTER TABLE
  "public"."marks"
ADD
  CONSTRAINT "marks_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."marks"
ADD
  CONSTRAINT "marks_relation_2" FOREIGN KEY ("lesson_id") REFERENCES "public"."lessons" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."marks"
ADD
  CONSTRAINT "marks_relation_3" FOREIGN KEY ("grade_id") REFERENCES "public"."grades" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."marks"
ADD
  CONSTRAINT "marks_relation_4" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
  
ALTER TABLE
  "public"."marks"
ADD
  CONSTRAINT "marks_relation_6" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE INDEX
  "marks_index_2" on "public"."marks"("user_id" ASC);

ALTER TABLE marks ADD CONSTRAINT lesson_mark_required_fields
CHECK(
  CASE WHEN type IN ('lesson_grade','not_done','notice_good','notice_neutral','notice_bad','absent', 'late')  
  THEN lesson_id is not NULL
  AND course is not NULL 
 END);

ALTER TABLE marks ADD CONSTRAINT grade_required_fields
CHECK(
  CASE WHEN type IN ('lesson_grade', 'course_grade', 'subject_grade')
  THEN grade_id is not NULL
END);

ALTER TABLE marks ADD CONSTRAINT course_grade_required_field
CHECK(
  CASE WHEN type = 'course_grade'
  THEN course is not NULL
END);