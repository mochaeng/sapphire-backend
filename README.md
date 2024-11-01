# Sapphire-backend

Golang backend for a social media app.

# Technologies used

-   Postgres for database
-   Redis for caching
-   [Chi](https://github.com/go-chi/chi) for routing
-   JWT for authentication
-   [Swaggo](https://github.com/swaggo/swag) for documentation
-   [Zap](https://github.com/uber-go/zap) for logging
-   [Go-mail](https://github.com/wneessen/go-mail) for e-mail sending

# Running

-   Start database:

```sh
docker compose up --build
```

-   Run migrations (You need to have [migrate](https://github.com/golang-migrate/migrate) installed as a CLI:

```sh
make migrate-up
```

-   Seed the database:

```sh
go run cmd/seed/main.go
```

-   Start application with [air](https://github.com/air-verse/air) (also need to be installed as a CLI)

```sh
air
```

# Bugs
