CREATE TABLE "messages" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "thread_id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "body" text NOT NULL,
    "type" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "messages"
    ADD CONSTRAINT "messages_relation_1" FOREIGN KEY ("thread_id") REFERENCES "threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "messages"
    ADD CONSTRAINT "messages_relation_2" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

CREATE INDEX messages_body_idx ON messages USING GIN (to_tsvector('simple', body));

---- create above / drop below ----

DROP TABLE "messages";