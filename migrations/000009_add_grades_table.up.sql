CREATE TABLE "public"."grades" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "identifier" varchar(3) NOT NULL,
    "value" integer NOT NULL
);

