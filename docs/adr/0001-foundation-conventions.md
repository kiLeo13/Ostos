# ADR 0001: Foundation conventions

- Status: Accepted
- Date: 2026-10-17

## Context

The approved `PLAN.md` supersedes earlier blueprint choices in `ARCHITECTURE.md`. The implementation needs one set of conventions before persistent entities and public mutations are introduced.

## Decision

- Use Go 1.27 with Echo v5 in one modular monolith.
- Use GORM for MySQL mapping and queries, explicit transaction boundaries, and an independent `golang-migrate` executable for schema changes. Application code must never call `AutoMigrate`.
- Use a Twitter-style 64-bit Snowflake identifier with a frozen epoch of `2026-01-01T00:00:00Z`, ten node bits, and twelve sequence bits. API IDs are decimal JSON strings; MySQL stores positive `BIGINT` values.
- Require a stable node ID in the inclusive range 0-1023 at API startup.
- Represent quantities as fixed thousandths backed by an integer and amounts as nonnegative integer cents. Floating point is forbidden in both paths.
- Emit UTC RFC 3339 timestamps and use `DATETIME(6)` in MySQL.
- Return the stable JSON error envelope defined by `PLAN.md`; parsed validation failures may include a `fields` object, while malformed JSON does not.

The epoch lasts beyond 2095 with 41 timestamp bits. It is immutable once any production ID is generated.

## Consequences

The service remains small and deployable as one binary, while migrations retain separate schema-administration credentials. Snowflake node assignment becomes an operational prerequisite. Clients must treat IDs and quantities as strings and cents as integers.
