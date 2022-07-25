# lavurso-backend

## depencencies
[pgx](https://github.com/jackc/pgx)

[chi](https://github.com/go-chi/chi)

[toml](https://github.com/BurntSushi/toml)

[bluemonday](https://github.com/microcosm-cc/bluemonday) (XSS protection)
    
 ## running
 1. copy `.makerc.example` to `.makerc`, edit `MIGRATE_DSN` to database connection string
 2. run `make db/migration_up`
 3. copy `config.toml.example` to `config.toml`, change values
 3. run `make api/run`
