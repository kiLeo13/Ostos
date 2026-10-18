# Project foundation

Status: implemented through plan step 4

## Boundaries

The backend is one Go modular monolith. Business packages own their future domain entities, policies, services, errors, and narrow repository interfaces. HTTP code depends on service-facing contracts; MySQL adapters depend inward on capability contracts. `cmd/api/main.go` is the only API composition root.

The accounts, Families, and shopping capabilities are implemented. Accounts use verified email/password authentication with short-lived JWT access tokens and rotating opaque refresh sessions. Families store ownership separately from Manager/Member rows. Shopping uses one implicit current list per Family and creates immutable trip history only when checked items are completed.

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

Migrations 2 through 4 create accounts, Families, current-list, and trip-history tables. `golang-migrate` owns its `schema_migrations` bookkeeping table. The API database user has data privileges only; only the migration process receives schema-administration credentials. GORM maps these tables but never migrates them.

`backend/compose.production.yml` is a hardened single-container baseline, not the completed public deployment. Origin CA TLS, port 443 publication, Cloudflare CIDR filtering, resource limits, health checks, and certificate mounts belong to production implementation step 6 and must be added together.

## Verification

Run the commands in `docs/runbooks/local-development.md`. The MySQL integration test is skipped unless `OSTOS_TEST_MYSQL_DSN` is set; the checked-in migration Compose project supplies a disposable real MySQL 8.4 instance and that variable.

Verification record, 2026-10-18: backend unit tests and `go vet` passed; all four migrations and the accounts, Families, and shopping repository suites passed against MySQL 8.4 in Compose; and the OpenAPI contract linted, regenerated, and type-checked cleanly. HeatWave compatibility remains a release gate because a local MySQL test cannot honestly stand in for the managed service.
