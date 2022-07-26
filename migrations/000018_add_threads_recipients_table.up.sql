create table if not exists
  "public"."threads_recipients" (
    "thread_id" INTEGER not null,
    "user_id" INTEGER not null,
    "group_id" INTEGER null,
    "read" BOOLEAN not null default FALSE
  );

alter table
  "public"."threads_recipients"
add
  constraint "threads_recipients_pkey" primary key ("thread_id", "user_id");

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