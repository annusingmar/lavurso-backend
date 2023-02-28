CREATE TABLE "threads" (
    "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id" integer NOT NULL,
    "title" text NOT NULL,
    "locked" boolean NOT NULL DEFAULT FALSE,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "threads"
    ADD CONSTRAINT "threads_relation_1" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

CREATE INDEX threads_title_idx ON threads USING GIN (to_tsvector('simple', title));

---- create above / drop below ----

DROP TABLE "threads";