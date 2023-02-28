CREATE TABLE "parents_children" (
    "parent_id" integer NOT NULL,
    "child_id" integer NOT NULL
);

ALTER TABLE "parents_children"
    ADD CONSTRAINT "parents_children_pkey" PRIMARY KEY ("parent_id", "child_id");

ALTER TABLE "parents_children"
    ADD CONSTRAINT "parents_children_relation_1" FOREIGN KEY ("parent_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "parents_children"
    ADD CONSTRAINT "parents_children_relation_2" FOREIGN KEY ("child_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "parents_children";