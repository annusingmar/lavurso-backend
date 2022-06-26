create table if not exists
  "public"."users_classes" ("user_id" INTEGER not null, "class_id" INTEGER not null);

alter table
  "public"."users_classes"
add
  constraint "users_classes_pkey" primary key ("user_id");

ALTER TABLE
  "public"."users_classes"
ADD
  CONSTRAINT "users_classes_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."users_classes"
ADD
  CONSTRAINT "users_classes_relation_2" FOREIGN KEY ("class_id") REFERENCES "public"."classes" ("id") ON UPDATE CASCADE ON DELETE CASCADE;