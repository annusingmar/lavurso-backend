create table if not exists
  "public"."absences_excuses" (
    "id" serial primary key,
    "absence_mark_id" INTEGER not null,
    "excuse" TEXT not null,
    "by" INTEGER not null,
    "at" TIMESTAMP not null default NOW()
  );

ALTER TABLE
  "public"."absences_excuses"
ADD
  CONSTRAINT "absences_excuses_relation_1" FOREIGN KEY ("absence_mark_id") REFERENCES "public"."marks" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."absences_excuses"
ADD
  CONSTRAINT "absences_excuses_relation_2" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;