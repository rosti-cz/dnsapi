# DNS API

Lightweight REST API for managing Bind9 zones and records across a primary/secondary nameserver setup.

The service runs in two modes:

- **Primary** — exposes the full zone/record CRUD API, persists data in SQLite, writes zone files and bind config to the local filesystem, and pushes updates to secondary instances via HTTP.
- **Secondary** — exposes only the bind-sync endpoints used by the primary to deliver zone files and named.conf. It writes received files to disk and reloads bind9.

Both modes are provided by the same binary; the runtime mode is controlled by `DNSAPI_MODE`.

## Quick start (Docker Compose)

```sh
cp docker-compose.yml.example docker-compose.yml   # adjust tokens and nameserver FQDNs
docker compose up -d
```

See [docker-compose.yml](docker-compose.yml) for a full example with comments explaining the rndc setup.

## Modes

### Primary mode

The primary instance:
- Serves the REST API (zones, records, commit, export/import).
- On commit: writes the zone file locally, pushes it to each secondary via `PUT /bind/zones/:domain`, regenerates and writes `named.conf.rosti` locally, and pushes it to each secondary via `PUT /bind/config`.
- Reloads bind9 via the configured reload command after updating local config/zones.
- 10 seconds after commit, forces an `rndc refresh` on itself and all secondaries.

### Secondary mode

The secondary instance:
- Exposes only the bind-sync endpoints (protected by the same API token as primary).
- On receiving a zone file or bind config via HTTP, writes it to disk and reloads bind9.

## Configuration

All settings are environment variables with the `DNSAPI_` prefix.

| Variable | Default | Description |
|---|---|---|
| `DNSAPI_MODE` | `primary` | Runtime mode: `primary` or `secondary` |
| `DNSAPI_API_TOKEN` | _(required)_ | Bearer token for all API calls |
| `DNSAPI_PORT` | `1323` | HTTP listen port (ignored when HTTPS is enabled) |
| `DNSAPI_PRIMARY_NAME_SERVER` | _(required in primary)_ | FQDN of the primary nameserver |
| `DNSAPI_NAME_SERVERS` | _(required in primary)_ | Comma-separated list of all nameservers (≥2) |
| `DNSAPI_PRIMARY_NAME_SERVER_IP` | _(auto-resolved)_ | Override for the primary NS IP — set this in containers/environments where the FQDN is not resolvable |
| `DNSAPI_SECONDARY_NAME_SERVER_IPS` | _(auto-resolved)_ | Override for the secondary NS IPs (zone `allow-transfer` / `masters`) — set this alongside the primary IP override |
| `DNSAPI_SECONDARY_INSTANCES` | _(empty)_ | Comma-separated HTTP URLs of secondary dnsapi instances |
| `DNSAPI_ABUSE_EMAIL` | _(required in primary)_ | SOA abuse email |
| `DNSAPI_DATABASE_PATH` | `gorm.sqlite` | SQLite database path (primary only) |
| `DNSAPI_BIND_RELOAD_COMMAND` | `systemctl reload bind9` | Shell command used to reload bind9 |
| `DNSAPI_BIND_REFRESH_COMMAND` | `rndc refresh` | Command prefix for zone refresh (zone name appended) |
| `DNSAPI_HTTPS` | `false` | Enable HTTPS with automatic Let's Encrypt certificate |
| `DNSAPI_PUBLIC_DOMAIN` | _(required when HTTPS)_ | Domain for the TLS certificate |
| `DNSAPI_ACME_EMAIL` | _(required when HTTPS)_ | Email for Let's Encrypt registration |
| `DNSAPI_SENTRY_DSN` | _(empty)_ | Sentry DSN for error reporting |
| `DNSAPI_SENTRY_ENV` | `dev` | Sentry environment tag |
| `DNSAPI_TIME_TO_REFRESH` | `300` | SOA refresh interval |
| `DNSAPI_TIME_TO_RETRY` | `180` | SOA retry interval |
| `DNSAPI_TIME_TO_EXPIRE` | `604800` | SOA expire value |
| `DNSAPI_MINIMAL_TTL` | `30` | SOA minimum TTL |
| `DNSAPI_TTL` | `3600` | Default record TTL |

## Installation (systemd)

Primary (runs on the same host as bind9):

```ini
[Unit]
Description=dnsapi — primary
After=network.target

[Service]
WorkingDirectory=/etc/bind
Environment=DNSAPI_MODE=primary
Environment=DNSAPI_API_TOKEN=<strong-secret>
Environment=DNSAPI_PRIMARY_NAME_SERVER=ns1.example.com
Environment=DNSAPI_NAME_SERVERS=ns1.example.com,ns2.example.com
Environment=DNSAPI_SECONDARY_INSTANCES=https://ns2.example.com:1323
Environment=DNSAPI_ABUSE_EMAIL=admin@example.com
Environment=DNSAPI_DATABASE_PATH=/var/lib/dnsapi/db.sqlite
Environment=DNSAPI_PORT=1323
ExecStart=/usr/local/bin/dnsapi
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Secondary (runs on the same host as the secondary bind9):

```ini
[Unit]
Description=dnsapi — secondary
After=network.target

[Service]
WorkingDirectory=/etc/bind
Environment=DNSAPI_MODE=secondary
Environment=DNSAPI_API_TOKEN=<same-strong-secret>
Environment=DNSAPI_PORT=1323
ExecStart=/usr/local/bin/dnsapi
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

## HTTPS / Let's Encrypt

Set `DNSAPI_HTTPS=true`, `DNSAPI_PUBLIC_DOMAIN=ns1.example.com`, and `DNSAPI_ACME_EMAIL=admin@example.com`.

The service starts on port 443 (TLS) and opens an HTTP-01 challenge listener on port 80.  
Certificates are cached in `/var/cache/dnsapi/autocert`.

When HTTPS is enabled, `DNSAPI_PORT` is ignored.

## API

Swagger UI is served at `/swagger/index.html` (primary mode only).

### Zones

| Method | Path | Description |
|---|---|---|
| `GET` | `/zones/` | List all zones |
| `GET` | `/zones/:zone_id` | Get a single zone |
| `POST` | `/zones/` | Create a zone |
| `PUT` | `/zones/:zone_id` | Update zone tags / abuse email |
| `DELETE` | `/zones/:zone_id` | Delete a zone |
| `PUT` | `/zones/:zone_id/commit` | Write zone to bind and sync to secondaries |

### Records

| Method | Path | Description |
|---|---|---|
| `GET` | `/zones/:zone_id/records/` | List records |
| `GET` | `/zones/:zone_id/records/:record_id` | Get a record |
| `POST` | `/zones/:zone_id/records/` | Create a record |
| `PUT` | `/zones/:zone_id/records/:record_id` | Update a record |
| `DELETE` | `/zones/:zone_id/records/:record_id` | Delete a record |

### Secondary sync endpoints (secondary mode only)

| Method | Path | Description |
|---|---|---|
| `PUT` | `/bind/config` | Write named.conf include file and reload bind9 |
| `PUT` | `/bind/zones/:domain` | Write a zone file |
| `DELETE` | `/bind/zones/:domain` | Delete a zone file |
| `POST` | `/bind/reload` | Reload bind9 |
| `POST` | `/bind/refresh/:domain` | Force zone refresh |

All endpoints require `Authorization: <token>` header.

## Database migrations

Versioned SQL migrations are embedded in the binary (`pkg/dnsapi/migrations/*.sql`) and applied automatically on startup. Migration state is tracked in the `schema_migrations` table.

## Development

```sh
go build ./...
go test ./...
make docs   # regenerate Swagger docs
```

