create table
  "public"."mark_history" (
    "id" serial primary key,
    "user_id" INTEGER not null,
    "mark_id" INTEGER not null,
    "lesson_id" INTEGER,
    "course" INTEGER,
    "journal_id" INTEGER,
    "grade_id" INTEGER,
    "subject_id" INTEGER,
    "comment" TEXT,
    "type" TEXT not null,
    "by" INTEGER not null,
    "at" TIMESTAMP default NOW()
  );

ALTER TABLE
  "public"."mark_history"
ADD
  CONSTRAINT "mark_history_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."mark_history"
ADD
  CONSTRAINT "mark_history_relation_2" FOREIGN KEY ("lesson_id") REFERENCES "public"."lessons" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."mark_history"
ADD
  CONSTRAINT "mark_history_relation_3" FOREIGN KEY ("grade_id") REFERENCES "public"."grades" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."mark_history"
ADD
  CONSTRAINT "mark_history_relation_4" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."mark_history"
ADD
  CONSTRAINT "mark_history_relation_5" FOREIGN KEY ("subject_id") REFERENCES "public"."subjects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."mark_history"
ADD
  CONSTRAINT "mark_history_relation_6" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."mark_history"
ADD
  CONSTRAINT "mark_history_relation_7" FOREIGN KEY ("mark_id") REFERENCES "public"."marks" ("id") ON UPDATE CASCADE ON DELETE CASCADE;