CREATE TABLE "log" (
    "user_id" integer,
    "session_id" integer,
    "method" text NOT NULL,
    "target" text NOT NULL,
    "ip" text NOT NULL,
    "response_code" integer NOT NULL,
    "duration" integer NOT NULL,
    "at" timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE "log"
    ADD CONSTRAINT "log_relation_1" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE "log"
    ADD CONSTRAINT "log_relation_2" FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

---- create above / drop below ----

DROP TABLE "log";