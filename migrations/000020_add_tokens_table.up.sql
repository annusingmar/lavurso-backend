create table
  "public"."tokens" (
    "id" serial primary key,
    "hash" BYTEA not null,
    "user_id" INTEGER not null,
    "type" TEXT not null,
    "expires" TIMESTAMP not null
  );

ALTER TABLE
  "public"."tokens"
ADD
  CONSTRAINT "tokens_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;