CREATE TABLE IF NOT EXISTS "public"."journals" (
    "id" serial PRIMARY KEY,
    "name" text NOT NULL,
    "teacher_id" integer NOT NULL,
    "subject_id" integer NOT NULL,
    "last_updated" timestamptz NOT NULL DEFAULT NOW(),
    "archived" boolean NOT NULL DEFAULT FALSE
);

ALTER TABLE "public"."journals"
    ADD CONSTRAINT "journals_relation_1" FOREIGN KEY ("teacher_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "public"."journals"
    ADD CONSTRAINT "journals_relation_2" FOREIGN KEY ("subject_id") REFERENCES "public"."subjects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

