create table
  "public"."done_assignments" (
    "user_id" INTEGER not null,
    "assignment_id" INTEGER not null
  );

alter table
  "public"."done_assignments"
add
  constraint "done_assignments_pkey" primary key ("user_id", "assignment_id");

ALTER TABLE
  "public"."done_assignments"
ADD
  CONSTRAINT "done_assignments_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE
  "public"."done_assignments"
ADD
  CONSTRAINT "done_assignments_relation_2" FOREIGN KEY ("assignment_id") REFERENCES "public"."assignments" ("id") ON UPDATE CASCADE ON DELETE CASCADE;