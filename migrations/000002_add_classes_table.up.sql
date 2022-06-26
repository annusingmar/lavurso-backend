CREATE TABLE IF NOT EXISTS
  "classes" (
    "id" serial primary key,
    "name" TEXT not null,
    "teacher_id" INTEGER not null,
    "archived" BOOLEAN not null DEFAULT false
);

ALTER TABLE
  "classes"
ADD
  CONSTRAINT "classes_relation_1" FOREIGN KEY ("teacher_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;