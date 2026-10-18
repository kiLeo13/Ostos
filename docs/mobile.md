# Mobile application

Status: plan step 5 implemented; physical-device journey validation pending

The app uses Expo SDK 58, Expo Router, strict TypeScript, TanStack Query, and the generated `@ostos/api-contract` client. SecureStore contains the access token, refresh token, and access expiry. SQLite key/value storage contains only the last successful Family/list snapshot, active Family ID, and `last_checked_at`.

Authentication includes registration, verification/resend, login, forgotten password, reset, profile/password changes, and current/all-session logout. Authenticated routes cover Family creation/selection, received invitations, the current list, completion decisions, history/detail, members/invitations, role changes, ownership transfer, Family exit/deletion, and account deletion.

Offline behavior is deliberately read-only. Device disconnection or an unreachable API displays the cached list and an accessible `Offline · last checked …` marker at the grocery tab. Mutation controls are disabled or rejected before success is shown. Launch, foreground, reconnect, pull-to-refresh, and successful mutations feed TanStack Query refresh behavior; logout clears credentials, query memory, and private snapshots.

Before calling the mobile phase release-ready, run these journeys on supported Android and iOS devices:

1. Register, follow the verification link, and log in after process restart.
2. Create a Family; invite by exact username and User ID; accept on another account.
3. Add, reorder, check/uncheck, price, and delete list items under each role.
4. Complete mixed purchased/keep/discard items and inspect immutable history.
5. Promote/demote, transfer ownership, remove/leave, and verify forbidden controls.
6. Launch in airplane mode after caching, kill/restart the process, reconnect, pull to refresh, and confirm no offline mutation is reported as successful.
7. Log out current/all sessions and confirm SecureStore plus cached private data are cleared.

These device journeys are documented but not claimed as executed in this repository-only implementation pass.

## Dependency audit note

As of 2026-10-18, `npm audit --omit=dev` reports high-severity findings in transitive Expo/Metro build-tool dependencies (`braces` and `node-forge`) and offers only a forced downgrade to Expo 44. The application is pinned to the current Expo SDK 58 compatibility set; do not apply that breaking downgrade. Re-run the audit on every Expo patch and upgrade when the upstream toolchain publishes compatible fixes.
