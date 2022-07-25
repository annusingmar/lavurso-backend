create table if not exists
  "public"."messages" (
    "id" serial primary key,
    "thread_id" INTEGER not null,
    "user_id" INTEGER not null,
    "body" TEXT not null,
    "type" TEXT not null,
    "created_at" TIMESTAMP not null default NOW(),
    "updated_at" TIMESTAMP not null default NOW(),
    "version" INTEGER not null default 1
  );

ALTER TABLE
  "public"."messages"
ADD
  CONSTRAINT "messages_relation_1" FOREIGN KEY ("thread_id") REFERENCES "public"."threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."messages"
ADD
  CONSTRAINT "messages_relation_2" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;