create table if not exists
  "public"."groups" ("id" serial primary key, "name" text not null, "archived" BOOLEAN not null DEFAULT false);