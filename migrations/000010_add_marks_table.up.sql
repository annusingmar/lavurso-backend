create table
  "public"."marks" (
    "id" serial primary key,
    "user_id" INTEGER not null,
    "lesson_id" INTEGER,
    "course" INTEGER,
    "journal_id" INTEGER,
    "grade_id" INTEGER,
    "subject_id" INTEGER,
    "comment" TEXT,
    "type" INTEGER not null,
    "current" BOOLEAN default true,
    "previous_ids" INTEGER [],
    "by" INTEGER not null,
    "at" TIMESTAMP default NOW()
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
  CONSTRAINT "marks_relation_4" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."marks"
ADD
  CONSTRAINT "marks_relation_5" FOREIGN KEY ("subject_id") REFERENCES "public"."subjects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."marks"
ADD
  CONSTRAINT "marks_relation_6" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;