# Accounts, Families, and shopping

Status: implemented by plan steps 2–4

Accounts require email verification before login. Access tokens are signed JWTs valid for 30 minutes; only hashed, rotating refresh tokens are persisted. Reuse revokes the affected replacement chain. Logout revokes the current refresh session, while logout-all and password reset revoke every session.

A Family has exactly one Owner in `families.owner_id`. Managers and Members live in `family_members`; API member collections synthesize the Owner. Invitations resolve only an exact username or User ID and never expose search suggestions. Creation, acceptance, ownership transfer, and the ten-Family limit use database transactions and row locks.

Each Family has one implicit current list. Every mutation locks the Family first, then rechecks membership, so completion and concurrent add/edit/delete/reorder operations share a stable lock order. Completion copies checked items to an immutable trip, deletes purchased items, retains unchecked `keep` items without history, and removes unchecked `discard` items. A completion with no checked items creates no empty trip.

Repository behavior is tested against MySQL 8.4 through the backend Compose integration profile. HeatWave compatibility remains a controlled pre-release smoke test.
