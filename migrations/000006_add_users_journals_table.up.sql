create table if not exists
  "public"."users_journals" (
    "user_id" INTEGER not null,
    "journal_id" INTEGER not null
  );

alter table
  "public"."users_journals"
add
  constraint "users_journals_pkey" primary key ("user_id", "journal_id");

ALTER TABLE
  "public"."users_journals"
ADD
  CONSTRAINT "users_journals_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."users_journals"
ADD
  CONSTRAINT "users_journals_relation_2" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;