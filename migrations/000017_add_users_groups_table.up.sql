create table if not exists
  "public"."users_groups" ("user_id" INTEGER not null, "group_id" INTEGER not null);

alter table
  "public"."users_groups"
add
  constraint "users_groups_pkey" primary key ("user_id", "group_id");

ALTER TABLE
  "public"."users_groups"
ADD
  CONSTRAINT "users_groups_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."users_groups"
ADD
  CONSTRAINT "users_groups_relation_2" FOREIGN KEY ("group_id") REFERENCES "public"."groups" ("id") ON UPDATE CASCADE ON DELETE CASCADE;