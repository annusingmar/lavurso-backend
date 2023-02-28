CREATE TABLE "teachers_classes" (
    "teacher_id" integer NOT NULL,
    "class_id" integer NOT NULL
);

ALTER TABLE "teachers_classes"
    ADD CONSTRAINT "teachers_classes_pkey" PRIMARY KEY ("teacher_id", "class_id");

ALTER TABLE "teachers_classes"
    ADD CONSTRAINT "teachers_classes_relation_1" FOREIGN KEY ("teacher_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "teachers_classes"
    ADD CONSTRAINT "teachers_classes_relation_2" FOREIGN KEY ("class_id") REFERENCES "classes" ("id") ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----

DROP TABLE "teachers_classes";