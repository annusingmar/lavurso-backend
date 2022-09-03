create table
  "public"."threads_read" (
    "thread_id" INTEGER not null,
    "user_id" INTEGER not null
  );

alter table
  "public"."threads_read"
add
  constraint "threads_read_pkey" primary key ("thread_id", "user_id");

ALTER TABLE
  "public"."threads_read"
ADD
  CONSTRAINT "threads_read_relation_1" FOREIGN KEY ("thread_id") REFERENCES "public"."threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."threads_read"
ADD
  CONSTRAINT "threads_read_relation_2" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;