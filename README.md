**Gator CLI**

Lightweight RSS aggregation CLI written in Go.

**Prerequisites**
- **Go**: Install Go (1.20+ recommended). See https://go.dev/doc/install for platform-specific instructions.
- **PostgreSQL**: A running Postgres server and a database for Gator. Ensure you can connect with a connection URL like `postgres://user:pass@localhost:5432/gator`.

**Install the gator CLI**
- **Build & install**: from the repository root run:

```bash
go install ./...
```

This installs the `gator` command into your `$GOPATH/bin` or `$GOBIN`.

**Database setup**
- Create a Postgres database for the app, for example:

```bash
createdb gator
```

- Run any provided migrations (this project includes SQL migration files under `sql/schema`). If you use goose (example):

```bash
goose postgres "postgres://user:pass@localhost:5432/gator" up
```

Replace the connection URL with your DB credentials.

**Config file**
- The app reads configuration via the internal `config` package. Provide a config containing at least the `DBUrl` and optionally defaults such as the current username. A minimal JSON example:

```json
{
  "DBUrl": "postgres://user:pass@localhost:5432/gator",
  "CurrentUsername": "alice"
}
```

- Save this file where your environment or the `config` package expects it (the repo uses `github.com/onurbilginnn/internal/config`). If unsure, create a local `config.json` and set an env var or consult the `config` package docs.

**Run the app**
- Run the CLI from the project root:

```bash
gator <command> [args]
```

Or run directly with `go run` while developing:

```bash
go run ./main.go <command> [args]
```

**Available commands**
- **login**: Set the current username in the config. Usage: `gator login username=<name>`
- **register**: Register a new user. Usage: `gator register username=<name>`
- **reset**: Reset users table. Usage: `gator reset`
- **users**: List users. Usage: `gator users`
- **agg**: Aggregate example that fetches a sample feed and prints items. Usage: `gator agg`
- **addfeed**: Add a feed and follow it. Requires authentication. Usage: `gator addfeed <name> <url>`
- **feeds**: List all feeds. Usage: `gator feeds`
- **follow**: Follow a feed (must be logged in). Usage: `gator follow feed_url=<url>`
- **following**: List feeds the current user follows. Usage: `gator following`
- **unfollow**: Unfollow a feed (must be logged in). Usage: `gator unfollow feed_url=<url>`
- **agg (daemon)**: Periodic aggregation (handlerAggregate) takes `time_between_reqs` (Go duration string like `1h` or `30m`). Example:

```bash
gator agg time_between_reqs=1h
```

This will run the aggregator on a ticker and print item titles to stdout.

**Notes & tips**
- The app uses generated DB helpers (sqlc). If you modify SQL files, run `sqlc generate` to refresh the Go code.
- When passing intervals to the aggregator or query logic use Go duration strings (`1h`, `30m`, `15m`).
- If you prefer to run only the binary without installing, run `go build` and execute the produced binary.

If you want, I can add a sample `config.json` to the repo or wire a `--config` flag to make config location explicit.
