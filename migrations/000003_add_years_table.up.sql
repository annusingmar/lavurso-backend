CREATE TABLE "public"."years" (
    "id" serial PRIMARY KEY,
    "display_name" text NOT NULL,
    "courses" integer NOT NULL,
    "current" boolean NOT NULL
);

ALTER TABLE years
    ADD CONSTRAINT at_least_one_course CHECK (courses > 0);

CREATE UNIQUE INDEX ON years (CURRENT)
WHERE
    CURRENT IS TRUE;

