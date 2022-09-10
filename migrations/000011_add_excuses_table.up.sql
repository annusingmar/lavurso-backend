CREATE TABLE IF NOT EXISTS "public"."excuses" (
    "mark_id" integer NOT NULL,
    "excuse" text NOT NULL,
    "by" integer NOT NULL,
    "at" timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE "public"."excuses"
    ADD CONSTRAINT "excuses_pkey" PRIMARY KEY ("mark_id");

ALTER TABLE "public"."excuses"
    ADD CONSTRAINT "excuses_relation_1" FOREIGN KEY ("mark_id") REFERENCES "public"."marks" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."excuses"
    ADD CONSTRAINT "excuses_relation_2" FOREIGN KEY ("by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

