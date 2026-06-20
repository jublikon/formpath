# Formpath

## Project Documentation

Project documentation lives under `docs/`:

- `docs/epics`: product slices and acceptance criteria
- `docs/adr`: architectural decisions
- `docs/changelog`: durable project memory for merged changes

Epics and changelogs have independent chronological numbering. Their
relationships are recorded explicitly in frontmatter:

```text
docs/epics/001-strava-activity-ingestion.md
docs/changelog/001-strava-activity-ingestion.md  -> related_epics: [epic-001]
docs/epics/002-training-overview.md
docs/changelog/003-training-overview.md          -> related_epics: [epic-002]
```

ADRs are linked the same way, because an epic or changelog can relate to
multiple architectural decisions.

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
- [MinIO](https://www.min.io/) API: http://localhost:9000
- [MinIO](https://www.min.io/) Console: http://localhost:9001

OAuth user tokens are stored in Postgres. Raw provider payloads are stored in [MinIO](https://www.min.io/), with object metadata and canonical records in Postgres.

When the backend runs inside Compose, `docker-compose.yml` overrides `DATABASE_URL` and `S3_ENDPOINT` to use service names. When the backend runs directly with `go run ./cmd/server`, the `.env.example` localhost values are the right shape.

When `DATABASE_URL` is configured, `S3_ENDPOINT`, `S3_ACCESS_KEY_ID`, and
`S3_SECRET_ACCESS_KEY` are also required. The backend validates this raw-storage
dependency during startup because canonical ingestion transforms only persisted
raw objects.
