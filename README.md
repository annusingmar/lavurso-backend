# lavurso-backend

## requirements

[Go](https://go.dev)

[PostgreSQL 15 (in beta)](https://www.postgresql.org) => data storage; the database must have extension `citext`

## depencencies

[pgx](https://github.com/jackc/pgx) => PostgreSQL database driver

[chi](https://github.com/go-chi/chi) => routing HTTP requests

[toml](https://github.com/BurntSushi/toml) => parsing config file

[bluemonday](https://github.com/microcosm-cc/bluemonday) => cross-site scripting (XSS) protection

## running

1.  clone this git repository: `git clone https://github.com/annusingmar/lavurso-backend.git`
2.  install the [`migrate`](https://github.com/golang-migrate/migrate) tool
3.  copy `.makerc.example` to `.makerc`, edit `MIGRATE_DSN` to database connection string
4.  run `make db/migration_up`
5.  copy `config.toml.example` to `config.toml`, change values
6.  run `make api/run`
