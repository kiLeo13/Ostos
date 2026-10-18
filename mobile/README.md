# Ostos mobile application

The initial client is an Expo SDK 58 and React Native application. It uses Expo
Router, strict TypeScript, TanStack Query, SecureStore for authentication
credentials, and SQLite-backed key/value storage for the last successful
Family/list snapshot.

## Local development

1. Copy `.env.example` to `.env.local` and set `EXPO_PUBLIC_API_URL` to an HTTPS
   API origin. Android emulators cannot reach a host process through
   `localhost`; use the host's reachable address instead.
2. Run `npm install`.
3. Run `npm start`, `npm run android`, or `npm run ios`.

The application never queues offline writes. When connectivity is absent, it
renders the last cached list read-only and identifies when that snapshot was
last refreshed.

## Validation

```text
npm run lint
npm run typecheck
npm test
npm run export:android
```
