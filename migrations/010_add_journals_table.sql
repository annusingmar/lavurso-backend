CREATE TABLE "journals" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name" text NOT NULL,
    "subject_id" integer NOT NULL,
    "year_id" integer NOT NULL,
    "last_updated" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "journals"
    ADD CONSTRAINT "journals_relation_1" FOREIGN KEY ("subject_id") REFERENCES "subjects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "journals"
    ADD CONSTRAINT "journals_relation_2" FOREIGN KEY ("year_id") REFERENCES "years" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

---- create above / drop below ----

DROP TABLE "journals";