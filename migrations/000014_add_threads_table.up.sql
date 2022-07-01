create table if not exists
  "public"."threads" (
    "id" serial primary key,
    "user_id" INTEGER not null,
    "title" TEXT not null,
    "locked" BOOLEAN not null default false,
    "created_at" TIMESTAMP not null
  );

ALTER TABLE
  "public"."threads"
ADD
  CONSTRAINT "threads_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;