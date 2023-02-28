CREATE TABLE "threads_recipients" (
    "thread_id" integer NOT NULL,
    "user_id" integer,
    "group_id" integer,
    UNIQUE NULLS NOT DISTINCT (thread_id, user_id, group_id)
);

ALTER TABLE "threads_recipients"
    ADD CONSTRAINT "threads_recipients_relation_1" FOREIGN KEY ("thread_id") REFERENCES "threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "threads_recipients"
    ADD CONSTRAINT "threads_recipients_relation_2" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "threads_recipients"
    ADD CONSTRAINT "threads_recipients_relation_3" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "threads_recipients";