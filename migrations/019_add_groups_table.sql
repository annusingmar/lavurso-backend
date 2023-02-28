CREATE TABLE "groups" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name" text NOT NULL,
    "archived" boolean NOT NULL DEFAULT FALSE
);

---- create above / drop below ----

DROP TABLE "groups";