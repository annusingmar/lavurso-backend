create table if not exists
  "public"."excuses" (
    "mark_id" INTEGER not null,
    "excuse" TEXT not null,
    "by" INTEGER not null,
    "at" TIMESTAMP not null default now()
  );

alter table
  "public"."excuses"
add
  constraint "excuses_pkey" primary key ("mark_id");

ALTER TABLE
  "public"."excuses"
ADD
  CONSTRAINT "excuses_relation_1" FOREIGN KEY ("mark_id") REFERENCES "public"."marks" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."excuses"
ADD
  CONSTRAINT "excuses_relation_2" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;