CREATE TABLE "public"."threads_read" (
    "thread_id" integer NOT NULL,
    "user_id" integer NOT NULL
);

ALTER TABLE "public"."threads_read"
    ADD CONSTRAINT "threads_read_pkey" PRIMARY KEY ("thread_id", "user_id");

ALTER TABLE "public"."threads_read"
    ADD CONSTRAINT "threads_read_relation_1" FOREIGN KEY ("thread_id") REFERENCES "public"."threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."threads_read"
    ADD CONSTRAINT "threads_read_relation_2" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

