CREATE TABLE IF NOT EXISTS "public"."groups" (
    "id" serial PRIMARY KEY,
    "name" text NOT NULL,
    "archived" boolean NOT NULL DEFAULT FALSE
);

