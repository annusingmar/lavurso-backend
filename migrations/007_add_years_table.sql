CREATE TABLE "years" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "display_name" text NOT NULL,
    "current" boolean NOT NULL
);

CREATE UNIQUE INDEX ON "years" (current)
WHERE
    current IS TRUE;

---- create above / drop below ----

DROP TABLE "years";