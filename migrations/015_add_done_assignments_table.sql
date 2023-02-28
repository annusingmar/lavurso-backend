CREATE TABLE "done_assignments" (
    "user_id" integer NOT NULL,
    "assignment_id" integer NOT NULL
);

ALTER TABLE "done_assignments"
    ADD CONSTRAINT "done_assignments_pkey" PRIMARY KEY ("user_id", "assignment_id");

ALTER TABLE "done_assignments"
    ADD CONSTRAINT "done_assignments_relation_1" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "done_assignments"
    ADD CONSTRAINT "done_assignments_relation_2" FOREIGN KEY ("assignment_id") REFERENCES "assignments" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "done_assignments";