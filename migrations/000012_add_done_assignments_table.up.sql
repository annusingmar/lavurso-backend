CREATE TABLE "public"."done_assignments" (
    "user_id" integer NOT NULL,
    "assignment_id" integer NOT NULL
);

ALTER TABLE "public"."done_assignments"
    ADD CONSTRAINT "done_assignments_pkey" PRIMARY KEY ("user_id", "assignment_id");

ALTER TABLE "public"."done_assignments"
    ADD CONSTRAINT "done_assignments_relation_1" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "public"."done_assignments"
    ADD CONSTRAINT "done_assignments_relation_2" FOREIGN KEY ("assignment_id") REFERENCES "public"."assignments" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

