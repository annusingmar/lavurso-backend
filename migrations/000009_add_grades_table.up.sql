CREATE TABLE "public"."grades" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "identifier" varchar(3) UNIQUE NOT NULL,
    "value" integer NOT NULL
);

