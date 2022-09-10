CREATE TABLE IF NOT EXISTS "public"."users_journals" (
    "user_id" integer NOT NULL,
    "journal_id" integer NOT NULL
);

ALTER TABLE "public"."users_journals"
    ADD CONSTRAINT "users_journals_pkey" PRIMARY KEY ("user_id", "journal_id");

ALTER TABLE "public"."users_journals"
    ADD CONSTRAINT "users_journals_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."users_journals"
    ADD CONSTRAINT "users_journals_relation_2" FOREIGN KEY ("journal_id") REFERENCES "public"."journals" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

