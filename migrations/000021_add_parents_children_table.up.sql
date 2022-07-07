create table if not exists
  "public"."parents_children" (
    "parent_id" INTEGER not null,
    "child_id" INTEGER not null
  );

alter table
  "public"."parents_children"
add
  constraint "parents_children_pkey" primary key ("parent_id", "child_id");

ALTER TABLE
  "public"."parents_children"
ADD
  CONSTRAINT "parents_children_relation_1" FOREIGN KEY ("parent_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."parents_children"
ADD
  CONSTRAINT "parents_children_relation_2" FOREIGN KEY ("child_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;