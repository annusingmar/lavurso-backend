create table if not exists
  "public"."users_threads" (
    "user_id" INTEGER not null,
    "thread_id" INTEGER not null
  );

alter table
  "public"."users_threads"
add
  constraint "users_threads_pkey" primary key ("user_id", "thread_id");

ALTER TABLE
  "public"."users_threads"
ADD
  CONSTRAINT "users_threads_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."users_threads"
ADD
  CONSTRAINT "users_threads_relation_2" FOREIGN KEY ("thread_id") REFERENCES "public"."threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;