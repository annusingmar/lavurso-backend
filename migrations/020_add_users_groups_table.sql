CREATE TABLE "users_groups" (
    "user_id" integer NOT NULL,
    "group_id" integer NOT NULL
);

ALTER TABLE "users_groups"
    ADD CONSTRAINT "users_groups_pkey" PRIMARY KEY ("user_id", "group_id");

ALTER TABLE "users_groups"
    ADD CONSTRAINT "users_groups_relation_1" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "users_groups"
    ADD CONSTRAINT "users_groups_relation_2" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "users_groups";