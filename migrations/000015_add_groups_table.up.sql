CREATE TABLE IF NOT EXISTS "public"."groups" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name" text NOT NULL,
    "archived" boolean NOT NULL DEFAULT FALSE
);

