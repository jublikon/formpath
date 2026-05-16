# Formpath

## Local Development

Prerequisite: install a local container runtime that supports Docker Compose, for example Docker Desktop, OrbStack, or Colima.

Create a local `.env` from `.env.example` and fill in the Strava app credentials:

```sh
cp .env.example .env
```

Start the full local stack:

```sh
docker compose up --build
```

Services:

- Backend: http://localhost:8080
- Postgres: localhost:5432
- MinIO API: http://localhost:9000
- MinIO Console: http://localhost:9001

OAuth user tokens are stored in Postgres. Raw provider payloads are stored in MinIO, with object metadata and canonical records in Postgres.

When the backend runs inside Compose, `docker-compose.yml` overrides `DATABASE_URL` and `S3_ENDPOINT` to use service names. When the backend runs directly with `go run ./cmd/server`, the `.env.example` localhost values are the right shape.
