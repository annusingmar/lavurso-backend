create table
  "public"."thread_log" (
    "action" TEXT not null,
    "target" INTEGER,
    "by" INTEGER not null,
    "at" TIMESTAMP not null
  );

ALTER TABLE
  "public"."thread_log"
ADD
  CONSTRAINT "thread_log_relation_1" FOREIGN KEY ("target") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."thread_log"
ADD
  CONSTRAINT "thread_log_relation_2" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;