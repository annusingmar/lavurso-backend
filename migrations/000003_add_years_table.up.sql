CREATE TABLE "public"."years" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "display_name" text NOT NULL,
    "current" boolean NOT NULL
);

CREATE UNIQUE INDEX ON years (CURRENT)
WHERE
    CURRENT IS TRUE;

