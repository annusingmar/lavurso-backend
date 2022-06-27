# lavurso-backend

## depencencies
[pgx](https://github.com/jackc/pgx)

[httprouter](https://github.com/julienschmidt/httprouter)
    
 ## runnning
 1. create `.makerc`, set `MIGRATE_DSN` to database connection string
 2. run `make db/migration_up`
 3. copy `config.toml.example` to `config.toml`, change values
 3. run `make api/run`
