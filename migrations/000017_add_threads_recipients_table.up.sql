create table if not exists
  "public"."threads_recipients" (
    "thread_id" INTEGER not null,
    "user_id" INTEGER,
    "group_id" INTEGER,
    unique NULLS NOT DISTINCT (thread_id, user_id, group_id)
  );

ALTER TABLE
  "public"."threads_recipients"
ADD
  CONSTRAINT "threads_recipients_relation_1" FOREIGN KEY ("thread_id") REFERENCES "public"."threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."threads_recipients"
ADD
  CONSTRAINT "threads_recipients_relation_2" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."threads_recipients"
ADD
  CONSTRAINT "threads_recipients_relation_3" FOREIGN KEY ("group_id") REFERENCES "public"."groups" ("id") ON UPDATE CASCADE ON DELETE CASCADE;