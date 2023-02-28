CREATE TABLE "classes_years" (
    "class_id" integer NOT NULL,
    "year_id" integer NOT NULL,
    "display_name" text NOT NULL
);

ALTER TABLE "classes_years"
    ADD CONSTRAINT "classes_years_pkey" PRIMARY KEY ("class_id", "year_id");

ALTER TABLE "classes_years"
    ADD CONSTRAINT "classes_years_relation_1" FOREIGN KEY ("class_id") REFERENCES "classes" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "classes_years"
    ADD CONSTRAINT "classes_years_relation_2" FOREIGN KEY ("year_id") REFERENCES "years" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "classes_years";