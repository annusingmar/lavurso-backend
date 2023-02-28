CREATE TABLE "classes" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name" text NOT NULL
);

---- create above / drop below ----

DROP TABLE "classes";