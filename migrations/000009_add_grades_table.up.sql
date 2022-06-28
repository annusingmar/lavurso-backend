create table
  "public"."grades" (
    "id" serial primary key,
    "identifier" VARCHAR(3) not null,
    "value" INTEGER not null
  );