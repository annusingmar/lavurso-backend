# lavurso-backend

## requirements

[Go](https://go.dev)
[PostgreSQL 15 (in beta)](https://www.postgresql.org)

## depencencies

[pgx](https://github.com/jackc/pgx) => PostgreSQL database driver

[chi](https://github.com/go-chi/chi) => routing HTTP requests

[toml](https://github.com/BurntSushi/toml) => parsing config file

[bluemonday](https://github.com/microcosm-cc/bluemonday) => cross-site scripting (XSS) protection

## running

1.  copy `.makerc.example` to `.makerc`, edit `MIGRATE_DSN` to database connection string
2.  run `make db/migration_up`
3.  copy `config.toml.example` to `config.toml`, change values
4.  run `make api/run`
