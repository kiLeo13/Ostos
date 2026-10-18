# Ostos contributor instructions

These rules apply to the entire repository. Read `ARCHITECTURE.md` before changing application behavior or infrastructure.

## Working agreement

- Restate the requested outcome, constraints, and acceptance criteria before non-trivial work.
- Plan non-trivial changes and obtain confirmation before implementation.
- Prefer small, reviewable changes and preserve unrelated work.
- Do not access or change a Git remote unless the user explicitly requests it.
- Do not perform destructive infrastructure, database, filesystem, or Git operations without explicit scope and a verified recovery path.
- Question requirements that weaken correctness, security, privacy, recoverability, or the Always Free constraint.

## Architecture boundaries

- Keep the backend a single Go modular monolith unless an ADR records measured evidence for another deployment boundary.
- Organize backend code by business capability (`accounts`, `families`, `shopping`, `catalog`, `retail`, `sync`), not global `handlers`, `services`, `repositories`, or `models` folders.
- Separate HTTP, application/domain, and MySQL concerns inside each capability.
- Domain/application code must not depend on Echo, SQL rows, AWS SES, Cloudflare, or OCI SDK types.
- Keep repository interfaces narrow and owned by their consumer. Do not introduce generic repositories.
- Go/Echo is the only application backend. Do not add Next.js server actions, a frontend BFF, or Cloudflare business logic that bypasses Go authorization.
- Keep Family authorization server-side and database-backed. UI permission checks are hints only.
- Preserve freeform shopping-list entries; catalog completeness must never block the list workflow.
- Preserve the distinction among category, brand, product, exact variant/GTIN, retailer listing, branch assortment, offer, list intent, and trip snapshot.
- Do not add microservices, Redis, queues, search clusters, WebSockets, or HeatWave acceleration without a measured need and an ADR.

## Data and API rules

- OpenAPI is the HTTP contract. Regenerate and verify server/client types whenever it changes.
- Use decimal-string Snowflake IDs externally, UTC timestamps, opaque cursor pagination, the stable JSON error envelope, record versions, and idempotency keys as specified in `PLAN.md` and `ARCHITECTURE.md`.
- Never use floating point for money or measured quantities.
- Mutations set desired state; never design retry-unsafe toggle endpoints.
- Enforce Family scope in every query and use composite constraints where practical.
- Multi-row domain transitions use a real `sql.Tx`, consistent lock order, context deadlines, and atomic commit/rollback.
- Migrations are forward-only and backward-compatible across a deployment. Do not assume schema downgrade is safe.
- Never test MySQL behavior with SQLite. Repository tests use a real MySQL engine.

## Security and privacy

- Never print, commit, or expose credentials, bearer tokens, invitation secrets, Terraform state, private shopping content, or environment files.
- Store invitation and refresh-like bearer secrets hashed.
- Verify authentication cryptographically and enforce authorization independently on every protected operation.
- Use exact CORS origins, bounded bodies, parameterized SQL, least-privilege DB credentials, and TLS.
- Private API responses are not cacheable. Product media must be validated, stripped of metadata, licensed, and served under immutable keys.
- Minimize personal data and maintain account export/deletion and retention behavior.

## Tests and validation

- Every behavior change requires appropriate tests. No exceptions.
- Critical permissions, ownership, invitation, money/quantity, trip-state, concurrency, and offline/idempotency branches require exhaustive coverage.
- Database interface changes require integration tests against real MySQL and a controlled compatibility smoke test against HeatWave before release.
- Web critical journeys use Playwright; mobile critical journeys include airplane mode, process death, restart, reconnect, and duplicate replay.
- Infrastructure changes require Terraform validation/plan review and Ansible syntax/idempotence checks.
- Backup changes are incomplete until a restore succeeds.
- A task is complete only when relevant tests are green and architecture/feature/runbook documentation reflects the result.

## Documentation and decisions

- Update `ARCHITECTURE.md` for system-wide behavior or constraints.
- Put consequential, durable tradeoffs in `docs/adr/` when that directory exists.
- Put operational procedures in `docs/runbooks/`.
- Keep route, permission, schema, deployment, and recovery documentation synchronized with implementation.
- Do not document aspirational behavior as if it is already implemented; label status clearly.

## Git conventions

- Do not initialize, commit, push, pull, rebase, merge, or open a pull request unless the requested task authorizes the relevant action.
- When commits are explicitly requested, make small coherent commits.
- Commit messages are normal text, start with a capital letter, and describe the change.
