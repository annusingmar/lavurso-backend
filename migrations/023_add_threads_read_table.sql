CREATE TABLE "threads_read" (
    "thread_id" integer NOT NULL,
    "user_id" integer NOT NULL
);

ALTER TABLE "threads_read"
    ADD CONSTRAINT "threads_read_pkey" PRIMARY KEY ("thread_id", "user_id");

ALTER TABLE "threads_read"
    ADD CONSTRAINT "threads_read_relation_1" FOREIGN KEY ("thread_id") REFERENCES "threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "threads_read"
    ADD CONSTRAINT "threads_read_relation_2" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "threads_read";