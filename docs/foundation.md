# Project foundation

Status: implemented by plan step 1

## Boundaries

The backend is one Go modular monolith. Business packages own their future domain entities, policies, services, errors, and narrow repository interfaces. HTTP code depends on service-facing contracts; MySQL adapters depend inward on capability contracts. `cmd/api/main.go` is the only API composition root.

The accounts, Families, and shopping packages are deliberately scaffolds in this step. Their schema and behavior belong to later plan steps; documenting fictional tables as implemented would be worse than leaving an honest boundary marker.

## Runtime conventions

- Go toolchain: 1.27.2
- HTTP: Echo v5
- Persistence: GORM with automatic migrations disabled
- Schema changes: the independent `migrations` module with embedded forward-only SQL
- IDs: positive Snowflakes with the frozen 2026-01-01 UTC epoch and a required 0-1023 process node ID
- Quantities: integer thousandths mapped to `DECIMAL(12,3)` and serialized as decimal strings
- Money: nonnegative integer BRL cents mapped to `BIGINT`
- Time: UTC internally, RFC 3339 over HTTP, `DATETIME(6)` in MySQL
- Units: `each`, `g`, `kg`, `ml`, and `l`

## HTTP safety

Mutation handlers must call the shared strict JSON decoder with a dedicated request DTO. The decoder requires `application/json`, rejects unknown fields and multiple JSON values, preserves numbers without converting through floating point, and returns all validation failures in one response. The router bounds bodies before decoding.

Errors always use the stable `errors` envelope. Internal causes are logged server-side and are never serialized. Authenticated responses added later must use `Cache-Control: private, no-store`.

## Database migration boundary

The initial migration establishes version 1 without inventing capability tables. `golang-migrate` owns its `schema_migrations` bookkeeping table. Capability steps add forward-only SQL and matching GORM models. The API database user has data privileges only; only the migration process receives schema-administration credentials.

`backend/compose.production.yml` is a hardened single-container baseline, not the completed public deployment. Origin CA TLS, port 443 publication, Cloudflare CIDR filtering, resource limits, health checks, and certificate mounts belong to production implementation step 6 and must be added together.

## Verification

Run the commands in `docs/runbooks/local-development.md`. The MySQL integration test is skipped unless `OSTOS_TEST_MYSQL_DSN` is set; the checked-in migration Compose project supplies a disposable real MySQL 8.4 instance and that variable.

Verification record, 2026-10-17: both modules built with Go 1.27.2; unit/race/vet/current-staticcheck scans passed; the migration ran twice idempotently on empty MySQL 8.4; GORM connected to the migrated schema without changing it; both container images built; and containerized liveness/readiness returned `200` with `{ "status": "ok" }`.
