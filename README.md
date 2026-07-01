# mz-monitoring

A website uptime monitoring service. Users register, add URLs of their sites, and the system periodically checks their availability in the background and stores the check history.

Written in Go using Clean Architecture (`domain -> usecase -> repository/delivery`).

## Features

- User registration and login (JWT auth, passwords hashed with bcrypt)
- CRUD for monitored sites (`targets`): create, list, delete
- Background scheduler that pushes check tasks to a queue at a configured interval (10 / 30 / 60 sec)
- Worker that actually makes the HTTP request, measures response time and status code
- Check history stored in PostgreSQL (`check_logs`) + a fast current status (online/offline) cached in Redis
- IP-based rate limiting via Redis (429 when exceeding 60 requests per minute)

## Stack

| Component        | Technology                      |
|-------------------|----------------------------------|
| Language          | Go                               |
| Database          | PostgreSQL                       |
| Cache / status     | Redis                            |
| Task queue         | RabbitMQ                         |
| Auth               | JWT (golang-jwt/jwt)             |
| Infrastructure      | Docker / Docker Compose          |

## Project structure

```
cmd/
  api/main.go         — HTTP API server
  scheduler/main.go   — check scheduler
  worker/main.go      — check executor (worker)
internal/
  domain/              — entities (User, Target, CheckLog) and repository interfaces
  usecase/             — business logic
  repository/
    postgres/          — repository implementations on top of Postgres
    redis/             — target status cache
    rabbitmq/           — task queue publisher
  delivery/http/       — HTTP handlers, JWT and rate-limit middleware
  worker/               — Scheduler and queue Consumer
pkg/
  jwt/                 — JWT generation and validation
  env/                 — simple .env loader
  rabbitmq/             — RabbitMQ connection and queue declaration
docker-compose.yaml
init.sql               — database schema
Dockerfile
```

## Database schema

**users** — `id, name, email (unique), password_hash, created_at`

**targets** — `id, user_id (FK), url, interval_sec (10/30/60), is_active`

**check_logs** — `id, target_id (FK), status_code, response_time_ms, is_up, checked_at`

## How it works

1. `scheduler` polls Postgres every 10 seconds for all active `targets` and, for the ones whose `interval_sec` is due, publishes a `{target_id, url}` message to the `monitoring_tasks` RabbitMQ queue.
2. `worker` consumes the queue, performs a `GET` request to the URL with a 5-second timeout, and measures the response time and status code.
3. The check result is written to Postgres (`check_logs`), and a lightweight online/offline status is cached in Redis so the API can return it without hitting Postgres.
4. `api` serves the sites CRUD and enriches the `GET /targets` response with the current status from Redis.

## Getting started

### 1. Environment variables

Create a `.env` file in the project root (used only for variable substitution in `docker-compose.yaml`, it is not copied into the image):

```env
JWT_SECRET=any-random-secret-string-here
```

### 2. Start everything with Docker Compose

```bash
docker compose up --build -d
```

This brings up: `postgres`, `redis`, `rabbitmq`, `api` (:8080), `scheduler`, `worker`. The DB schema is applied automatically from `init.sql` on the first Postgres container start.

Check status:

```bash
docker compose ps
docker compose logs -f api
```

RabbitMQ management UI: http://localhost:15672 (guest / guest)

If you need to re-apply the schema (e.g. after editing `init.sql`):

```bash
docker compose down -v
docker compose up --build -d
```

## API

Base URL: `http://localhost:8080`

### `POST /auth/register`

```json
{
  "name": "Test",
  "email": "test@test.com",
  "password": "12345678"
}
```

→ `201 Created`, `{ "status": "success", "message": "...", "id": 1 }`

### `POST /auth/login`

```json
{
  "email": "test@test.com",
  "password": "12345678"
}
```

→ `200 OK`, `{ "token": "<jwt>" }`

From here on, pass the token in the `Authorization: Bearer <jwt>` header for all requests below.

### `POST /targets`

```json
{
  "url": "https://google.com",
  "interval_sec": 10
}
```

`interval_sec` — only `10`, `30`, or `60`.

### `GET /targets`

Returns the user's sites with their current status:

```json
[
  {
    "id": 1,
    "user_id": 1,
    "url": "https://google.com",
    "interval_sec": 10,
    "is_active": true,
    "is_online": true
  }
]
```

### `DELETE /targets/:id`

Deletes a site (only your own — scoped by `user_id` from the token).

### Rate limiting

Any IP sending more than 60 requests per minute gets `429 Too Many Requests`.

## Quick check with curl

```bash
# register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@test.com","password":"12345678"}'

# login
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"12345678"}' | jq -r .token)

# create a target
curl -X POST http://localhost:8080/targets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://google.com","interval_sec":10}'

# list targets (is_online will appear after 10-20 sec)
curl http://localhost:8080/targets -H "Authorization: Bearer $TOKEN"
```

## Inspecting the database

```bash
docker exec -it mz_postgres psql -U postgres -d mz_monitoring
```

Then inside `psql`:

```sql
\dt                          -- list tables
\d targets                   -- describe a table's columns
SELECT * FROM check_logs ORDER BY checked_at DESC LIMIT 20;
```

Or connect from a GUI client (TablePlus, DBeaver, etc.) to `localhost:5432`, user `postgres`, password `mysecretpassword`, database `mz_monitoring`.

Redis status keys:

```bash
docker exec -it mz_redis redis-cli
KEYS target:status:*
```


## Roadmap / possible improvements

- [ ] Graceful shutdown with proper `ack`/`nack` in the RabbitMQ consumer
- [ ] Pagination for `GET /targets`
- [ ] Check history endpoint (`GET /targets/:id/logs`)
- [ ] Metrics (Prometheus) and a health-check endpoint
- [ ] Tests (unit tests for the usecase layer, integration tests for repositories)