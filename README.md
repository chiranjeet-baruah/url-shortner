# url-shortener

A minimal URL shortener built with Go, PostgreSQL, and Redis.

## Stack

- **Go** — standard library HTTP server, no frameworks
- **PostgreSQL** — persistent storage
- **Redis** — URL lookup cache & async click buffering

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/shorten` | Create a short URL |
| `GET` | `/r/{shortCode}` | Redirect to the original URL |
| `GET` | `/api/v1/stats/{shortCode}` | Get click stats for a URL |
| `GET` | `/` | Web UI |

### Shorten a URL

```bash
curl -s -X POST http://localhost:8080/api/v1/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://example.com"}' | jq
```

## Run

```bash
docker compose up --build
```

The app is available at **<http://localhost:8080>**.

## Project structure

```
cmd/server/          Entry point
internal/v1/shortener/
  adapter/driven/    PostgreSQL, Redis, analytics adapters
  adapter/driver/    HTTP router & handlers
  domain/            Core types & port interfaces
  dto/               Request/response structs
  service/           Business logic
  constant/          Constants
web/                 Embedded static web UI
```
