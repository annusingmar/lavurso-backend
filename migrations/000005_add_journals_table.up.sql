create table if not exists
  "public"."journals" (
    "id" serial primary key,
    "name" TEXT not null,
    "teacher_id" INTEGER not null,
    "subject_id" INTEGER not null,
    "archived" BOOLEAN not null DEFAULT false
  );

ALTER TABLE
  "public"."journals"
ADD
  CONSTRAINT "journals_relation_1" FOREIGN KEY ("teacher_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE
  "public"."journals"
ADD
  CONSTRAINT "journals_relation_2" FOREIGN KEY ("subject_id") REFERENCES "public"."subjects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;