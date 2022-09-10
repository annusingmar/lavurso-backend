CREATE TABLE IF NOT EXISTS "public"."users_groups" (
    "user_id" integer NOT NULL,
    "group_id" integer NOT NULL
);

ALTER TABLE "public"."users_groups"
    ADD CONSTRAINT "users_groups_pkey" PRIMARY KEY ("user_id", "group_id");

ALTER TABLE "public"."users_groups"
    ADD CONSTRAINT "users_groups_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."users_groups"
    ADD CONSTRAINT "users_groups_relation_2" FOREIGN KEY ("group_id") REFERENCES "public"."groups" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

