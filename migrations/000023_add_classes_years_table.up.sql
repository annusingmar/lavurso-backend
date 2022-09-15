CREATE TABLE "public"."classes_years" (
    "class_id" integer NOT NULL,
    "year_id" integer NOT NULL,
    "display_name" text NOT NULL
);

ALTER TABLE "public"."classes_years"
    ADD CONSTRAINT "classes_years_pkey" PRIMARY KEY ("class_id", "year_id");

ALTER TABLE "public"."classes_years"
    ADD CONSTRAINT "classes_years_relation_1" FOREIGN KEY ("class_id") REFERENCES "public"."classes" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."classes_years"
    ADD CONSTRAINT "classes_years_relation_2" FOREIGN KEY ("year_id") REFERENCES "public"."years" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

