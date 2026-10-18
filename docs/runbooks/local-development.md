# Local development

## Prerequisites

- Go 1.27.2 or a Go toolchain capable of downloading it
- Docker with Compose
- Node.js 24 and npm for OpenAPI and mobile development

## Start the API and MySQL

From `backend/` run:

```sh
docker compose up --build
```

The Compose project starts MySQL, runs forward migrations with schema-administration credentials, then starts the API with a lower-privilege application account. The API listens on `http://localhost:8080`; liveness and readiness are available at `/health/live` and `/health/ready`.

The checked-in passwords are local-only Compose defaults. Never reuse them outside a disposable local database.

## Validate the foundation

```sh
cd backend
go test ./...
go vet ./...

cd ../migrations
go test ./...
go vet ./...

cd ../api
npm ci
npm run lint
npm run generate
npm run check:generated
npm run typecheck
npm audit --audit-level=moderate
```

Run real-MySQL integration tests through `migrations/compose.yml`; SQLite is intentionally unsupported.

```sh
cd migrations
docker compose up --abort-on-container-exit --exit-code-from integration integration
docker compose stop

cd ../backend
docker compose --profile test up --build --abort-on-container-exit --exit-code-from integration integration
docker compose stop
```

Use `go run honnef.co/go/tools/cmd/staticcheck@latest ./...` and `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` in each Go module so the analysis toolchain supports Go 1.27.

## Run the mobile application

Copy `mobile/.env.example` to the ignored `mobile/.env.local` and set `EXPO_PUBLIC_API_URL`. Use an address reachable by the device or emulator; Android cannot reach the host API through its own `localhost`.

```sh
cd mobile
npm ci
npm run lint
npm run typecheck
npm test
npm start
```

`npm test` covers token persistence/rotation, rejected-refresh cleanup, stable API error mapping, offline mutation rejection, connectivity states, and the accessible stale-cache warning. Use `npm run export:android` as the Metro bundle smoke test before a mobile release.
