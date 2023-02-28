CREATE TABLE "excuses" (
    "mark_id" integer NOT NULL,
    "excuse" text NOT NULL,
    "user_id" integer NOT NULL,
    "at" timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE "excuses"
    ADD CONSTRAINT "excuses_pkey" PRIMARY KEY ("mark_id");

ALTER TABLE "excuses"
    ADD CONSTRAINT "excuses_relation_1" FOREIGN KEY ("mark_id") REFERENCES "marks" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "excuses"
    ADD CONSTRAINT "excuses_relation_2" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

---- create above / drop below ----

DROP TABLE "excuses";