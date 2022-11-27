CREATE TABLE IF NOT EXISTS "classes" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name" text NOT NULL
);

ALTER TABLE "classes"
    ADD CONSTRAINT "classes_relation_1" FOREIGN KEY ("teacher_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

