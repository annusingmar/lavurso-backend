CREATE TABLE IF NOT EXISTS "classes" (
    "id" serial PRIMARY KEY,
    "name" text NOT NULL,
    "teacher_id" integer NOT NULL,
    "archived" boolean NOT NULL DEFAULT FALSE
);

ALTER TABLE "classes"
    ADD CONSTRAINT "classes_relation_1" FOREIGN KEY ("teacher_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

