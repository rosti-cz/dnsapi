# AGENTS.md

Operational notes for AI agents (and humans) working on this repository.

## Project

- **Name**: `dnsapi` — REST API + web UI for managing BIND9 DNS zones.
- **Module path**: `github.com/by-cx/dnsapi`
- **Root**: `~/co/rosti/dnsapi`
- **Language**: Go 1.21. Requires CGO (SQLite via jinzhu/gorm v1 + go-sqlite3).
- **HTTP framework**: Echo v3 (`labstack/echo v3.3.10`)
- **ORM**: jinzhu/gorm v1.9.16 with SQLite3. `CGO_CFLAGS="-D_LARGEFILE64_SOURCE"` is set in the Dockerfile for large-file support.
- **Swagger**: swag v1.16.6 + http-swagger/v2 v2.0.2. Regenerate with `make docs`.
- **UI**: single embedded HTML file (`pkg/dnsapi/ui/index.html`) using Alpine.js v3 + TailwindCSS CDN (no build step). Served at `GET /ui` without auth.

## Repo layout

```
main.go                   # entry point — calls pkg/dnsapi.Run()
pkg/dnsapi/
  app.go                  # Echo setup, route registration, embed directive for UI
  types.go                # Zone and Record GORM model structs
  api_models.go           # Request/response DTOs
  processors.go           # Business logic (NewZone, UpdateZone, Commit, …)
  handlers.go             # Primary-mode HTTP handlers with Swagger annotations
  dns_check.go            # GET /zones/:zone_id/test — live NS record comparison
  zone_status.go          # GET /zones/:zone_id/status — commit/NS config status
  middlewares.go          # Auth and Sentry middleware; /ui is in skipPaths (no auth)
  migrations/             # Numbered SQL migration files (applied on startup)
  ui/index.html           # Embedded single-file SPA
docs/                     # Generated Swagger docs (do not edit by hand)
Makefile                  # Build, test, docs targets
docker-compose.yml        # Dev environment: bind-primary, dnsapi-primary, bind-secondary, dnsapi-secondary
Dockerfile                # Multi-stage build; sets CGO_CFLAGS for sqlite3
```

## Build and test

```sh
# Run tests
make test           # go test -run '' -v

# Vet
make vet

# Regenerate Swagger docs after changing handler annotations
make docs

# Build release binaries (linux amd64/arm/arm64)
make build

# Quick local compile check
go build ./...
```

## Dev environment (podman-compose)

The `docker-compose.yml` spins up 4 containers:

- `bind-primary` — authoritative BIND9 master on `localhost:5352`
- `dnsapi-primary` — dnsapi in primary mode on `localhost:1323` (API token: `abcd`)
- `bind-secondary` — BIND9 slave on `localhost:5354`
- `dnsapi-secondary` — dnsapi in secondary mode on `localhost:1324`

The SQLite database is stored in the `dnsapi-data` named volume, mounted at `/data/dnsapi.sqlite` inside the container.

```sh
# Build images
podman-compose build --no-cache

# Start (keep existing data)
podman-compose down
podman-compose up -d

# View logs
podman-compose logs -f dnsapi-primary

# Rebuild without wiping data — do NOT pass -v to down
podman-compose build --no-cache && podman-compose down && podman-compose up -d
```

> **Warning**: `podman-compose down -v` deletes ALL named volumes including the database.
> Never use `-v` unless you explicitly want to reset all data.

## Testing the API and UI

- **Web UI**: `http://localhost:1323/ui` (no auth required in browser; token is stored in localStorage)
- **Swagger UI**: `http://localhost:1323/swagger/index.html`
- **API token for dev**: `abcd` — pass as `Authorization: abcd` header
- **HTTP request examples**: `requests.http` in repo root (works with VS Code REST Client)

Quick smoke test:
```sh
curl -s -H "Authorization: abcd" http://localhost:1323/zones/ | python3 -m json.tool
```

## JSON field naming

The API returns lowercase snake_case JSON keys matching the `json:` struct tags in `types.go` and `api_models.go`. The UI (`ui/index.html`) must use these lowercase names (e.g. `zone.id`, `zone.domain`, `zone.committed_serial`, `rec.name`, `rec.ttl`) — **not** Go field names like `zone.ID` or `zone.Domain`.

## Architecture notes

- All business logic lives in `processors.go`. Handlers and the UI call processors; they do not contain logic themselves.
- `Commit()` in processors writes the zone file to disk, triggers `rndc reload`, and saves `committed_serial` equal to `Serial` in the DB. The UI shows an "uncommitted" badge when `zone.serial !== zone.committed_serial`.
- Migrations in `pkg/dnsapi/migrations/` are numbered and applied automatically on startup. To add a migration, create the next numbered `.sql` file.
- The `owner` field on zones is populated either via the API or automatically by migration `005_migrate_owner_from_tags.sql` which extracts `owner:NNNN` tags and writes `companyNNNN` into the `owner` column.
