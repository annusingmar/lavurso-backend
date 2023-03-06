ALTER TABLE "users" ADD "totp_enabled" boolean NOT NULL DEFAULT FALSE;

ALTER TABLE "users" ADD "totp_secret" text;

---- create above / drop below ----

ALTER TABLE "users" DROP "totp_enabled";

ALTER TABLE "users" DROP "totp_secret";