CREATE TABLE "grades" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "identifier" varchar(3) UNIQUE NOT NULL,
    "value" integer NOT NULL
);

---- create above / drop below ----

DROP TABLE "grades";