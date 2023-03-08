ALTER TABLE "log" RENAME TO "logs";

ALTER TABLE "logs" ADD COLUMN "id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

---- create above / drop below ----

ALTER TABLE "logs" DROP COLUMN "id";

ALTER TABLE "logs" RENAME TO "log";