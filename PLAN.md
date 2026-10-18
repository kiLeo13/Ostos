# Ostos implementation plan

Status: approved implementation baseline  
Last reconciled: 2026-10-17  
Initial delivery: React Native application and Go API  
Near-future delivery: complete authenticated web application

Ostos is a small, real grocery-collaboration application for a small userbase in one city. A user creates or joins Families, each Family has one shared grocery list, and participants plan purchases together. Completing the list records purchased items in immutable history, retains unchecked items selected for the next purchase, and deletes unchecked items that are no longer wanted.

The initial product includes accounts, email verification, Families, hierarchy, invitations, the current grocery list, purchase completion, history, and a read-only offline cache. It intentionally excludes a product catalog, supermarket API integration, price comparison, images, notifications, offline mutations, WebSockets, a search server, Prometheus, SSR, and application-managed database backups. Those exclusions keep the implementation proportional to its audience and the OCI Always Free budget.

This document is the implementation authority for Ostos. It records later approved decisions when it conflicts with `ARCHITECTURE.md`; that document must be reconciled before application implementation begins.

## 1. Build the project foundation

1. Implement the backend as one Go modular monolith. Use the current supported Go release, Echo v5 for HTTP, GORM for MySQL access and explicit transactions, and MySQL HeatWave `MySQL.Free` as the transactional database. GORM maps and queries entities but never runs migrations.
2. Use this repository layout:

   ```text
   /
     AGENTS.md
     ARCHITECTURE.md
     PLAN.md

     api/
       openapi.yaml

     backend/
       go.mod
       go.sum
       Dockerfile
       compose.yml
       compose.production.yml
       cmd/
         api/
           main.go
       internal/
         accounts/
         families/
         shopping/
         apierrors/
         config/
         platform/
         repository/
           mysql/
         transport/
           http/
             handler/
             middleware/
             request/
             response/
             router/
             validation/

     migrations/
       go.mod
       go.sum
       Dockerfile
       compose.yml
       cmd/
         migrate/
           main.go
       sql/

     mobile/
     web/                    # Created only when step 7 begins.

     infra/
       terraform/
         bootstrap/
         environments/
           production/
         modules/
       ansible/
         inventories/
         playbooks/
         roles/

     docs/
       adr/
       runbooks/
   ```

3. Organize business code by capability. Each capability owns its entities, use cases, policies, errors, and the narrow repository interface its services consume:

   ```text
   backend/internal/accounts/
     user.go
     credential.go
     session.go
     repository.go
     service.go
     errors.go

   backend/internal/families/
     family.go
     member.go
     invitation.go
     repository.go
     service.go
     policy.go
     errors.go

   backend/internal/shopping/
     list_item.go
     trip.go
     trip_item.go
     repository.go
     service.go
     policy.go
     errors.go
   ```

4. Keep concrete database adapters together for this project size. Explicit filenames and constructor names prevent the ambiguity created by many identically named `mysql` packages:

   ```text
   backend/internal/repository/mysql/
     accounts.go
     families.go
     shopping.go
     email_outbox.go
     audit.go
     transactions.go
   ```

   ```go
   accountRepository := mysql.NewAccountRepository(db)
   familyRepository := mysql.NewFamilyRepository(db)
   shoppingRepository := mysql.NewShoppingRepository(db)
   ```

5. Keep dependency direction explicit:
   1. Capability packages define their own repository interfaces and cannot import Echo or the MySQL adapter.
   2. The MySQL package imports capability packages and implements their interfaces.
   3. HTTP handlers depend on small service interfaces and translate transport DTOs to application inputs.
   4. `cmd/api/main.go` is the composition root that creates concrete adapters and wires constructors.
   5. Constructors accept interfaces and return concrete implementations, for example `NewService(repository Repository, ids IDGenerator, clock Clock) *Service`.
   6. Persistent capability entities may contain GORM tags, but handlers never bind requests into them or serialize them directly.
6. Establish these cross-cutting data conventions before creating entities:
   1. Generate every platform entity ID with a Twitter-style Snowflake implementation, store it as a positive MySQL `BIGINT`, and expose it as a decimal JSON string.
   2. Freeze the Snowflake epoch before the first production record. Assign each API process a stable node ID from `0` through `1023` and fail startup if it is absent or invalid.
   3. Use `snake_case` JSON, UTC internally, RFC 3339 API timestamps, and `DATETIME(6)` database timestamps.
   4. Give mutable entities `created_at` and `updated_at`.
   5. Decode quantities through `json.Number` or a fixed-decimal type and store them as `DECIMAL(12,3)`; binary floating point must not enter quantity or money calculations.
   6. Initially support `each`, `g`, `kg`, `ml`, and `l` units.
   7. Store amounts as nonnegative integer cents in `BIGINT`. The deployment currency is BRL, so a per-row currency column is unnecessary until multi-currency support exists.
7. Make `api/openapi.yaml` the HTTP source of truth. Generate transport types where useful and one TypeScript client for mobile and web. A contract change is complete only after OpenAPI linting, compatibility review, regeneration, and a clean generated-code diff.
8. Bind every JSON body into a dedicated request DTO and validate it with one application-wide `go-playground/validator/v10` instance. This includes `UserCreateRequest`, account commands, Family commands, invitation and membership commands, list-item commands, ordering, and list completion.
   1. Use built-in tags such as `required`, `min`, `max`, `email`, `oneof`, `gt`, and `omitempty` for structural rules.
   2. Register struct-level validators for mutually exclusive or related fields.
   3. Register a reusable `username` validator that accepts only 2–32 lowercase ASCII letters, digits, hyphens, and underscores.

      ```go
      Username string `json:"username" validate:"required,min=2,max=32,username"`
      ```

   4. Collect every field failure in the same validation pass. Database uniqueness, authorization, ownership, and membership limits remain service/database concerns.
9. Return JSON errors through one stable envelope. Parsed requests with invalid fields include every field failure:

   ```json
   {
     "errors": {
       "code": "INVALID_FORM_BODY",
       "message": "One or more fields are invalid",
       "fields": {
         "email": "Enter a valid email address",
         "username": "Use 2-32 lowercase letters, digits, hyphens, or underscores"
       }
     }
   }
   ```

   Malformed JSON has no `fields` member because no field-level validation occurred:

   ```json
   {
     "errors": {
       "code": "MALFORMED_JSON",
       "message": "The request body must contain valid JSON"
     }
   }
   ```

   Use `400` for malformed semantics, `401` for invalid authentication, `403` for insufficient permission, `404` for unavailable resources, `409` for state/uniqueness conflicts, `413` for oversized bodies, `415` for unsupported media types, `422` for field validation, `429` for rate limits, and generic `500`/`503` responses for internal/dependency failure.
10. Build migrations as an independent root Go module before implementing repositories. Its only executable is `migrations/cmd/migrate/main.go`; it loads configuration, opens MySQL with bounded timeouts, invokes the selected versioned migration library, and exits nonzero on failure. Store forward-oriented SQL under `migrations/sql` and use separate schema-administration credentials that the API never receives.
11. Use Docker and Compose for local development, migration execution, real-MySQL integration tests, and the single production API container. Keep production lightweight: MySQL stays managed, and no reverse proxy, search engine, metrics server, or auxiliary application container runs on the VM.
12. Finish this foundation when:
    1. Every Go module builds and common formatting/static-analysis commands pass.
    2. The migration module can migrate an empty real MySQL instance and GORM mappings match the resulting schema without `AutoMigrate`.
    3. Request binding rejects unknown mutation fields, enforces body limits and content type, and produces the documented error shapes.
    4. Snowflake, decimal quantity, cents, timestamp, validation, and error-mapping tests pass.

## 2. Implement accounts and authentication

1. Create the account schema with these entities:
   1. `users`: Snowflake `id`, canonical unique `email`, validated unique lowercase `username`, `display_name`, nullable `email_verified_at`, `created_at`, `updated_at`, and optional `deleted_at`. Store the canonical values directly; do not duplicate them into `email_normalized` or `username_normalized` columns.
   2. `user_credentials`: Snowflake `id`, unique `user_id`, Argon2id `password_hash`, `password_changed_at`, `created_at`, and `updated_at`.
   3. `auth_sessions`: Snowflake `id`, `user_id`, unique refresh-token hash, optional bounded device name, `created_at`, `last_used_at`, `expires_at`, nullable `revoked_at`, and nullable `replaced_by_session_id` for rotation.
   4. `account_tokens`: Snowflake `id`, `user_id`, purpose (`verify_email` or `reset_password`), unique bearer-token hash, `expires_at`, nullable `used_at`, and `created_at`.
   5. `email_outbox`: Snowflake `id`, `user_id`, account-email template, recipient, safe template data, encrypted short-lived token payload, delivery state, attempt count, next-attempt time, nullable sent time, `created_at`, and `updated_at`. Keep the encryption key outside MySQL, clear the encrypted token after delivery, and retain only the single-use hash in `account_tokens`.
2. Publish these account and authentication routes:

   ```text
   POST   /v1/users
   GET    /v1/users/me
   PATCH  /v1/users/me
   PUT    /v1/users/me/password
   DELETE /v1/users/me

   POST /v1/auth/login
   POST /v1/auth/refresh
   POST /v1/auth/logout
   POST /v1/auth/logout-all
   POST /v1/auth/verify-email
   POST /v1/auth/resend-verification
   POST /v1/auth/forgot-password
   POST /v1/auth/reset-password
   ```

3. Implement `POST /v1/users` as one transaction that canonicalizes the email, validates all input, hashes the password with Argon2id, inserts the User and credential, creates a hashed verification token, and enqueues its email. Return `201 Created`, a `Location` header, and the User resource:

   ```json
   {
     "id": "241895317264384000",
     "email": "user@example.com",
     "username": "leonardo",
     "display_name": "Leonardo",
     "email_verified": false,
     "created_at": "2026-10-17T12:00:00Z",
     "updated_at": "2026-10-17T12:00:00Z"
   }
   ```

4. Deliver verification and password-reset emails through AWS SES. Tokens are high-entropy, expiring, single-use values whose hashes are stored in MySQL. Configure the sending domain with SES verification, DKIM, SPF, and DMARC, and leave the SES sandbox before sending to arbitrary recipients.
5. Require email verification before login. Successful login and refresh return:
   1. A signed access JWT valid for 30 minutes.
   2. A random opaque refresh token backed by `auth_sessions`.
6. Include only the subject User ID, session ID, issuer, audience, issued-at, expiry, and unique token ID in access JWTs. Family IDs and roles are database state and never JWT claims.
7. Store only refresh-token hashes. Rotate the token on every refresh; reuse of a replaced token revokes its affected session chain. Normal logout revokes the current session, while logout-all and password reset revoke every session owned by the User. Access JWTs are not stored or denylisted, so an issued token may remain cryptographically valid for at most 30 minutes after logout.
8. Enforce current Family authorization from MySQL on every Family operation. Removing a member therefore takes effect immediately even while their account JWT remains valid.
9. Let `/v1/users/me` return profile information, not Family IDs. `GET /v1/families` returns the authorized Family summaries directly and avoids a client-side N+1 fetch.
10. Require recent password confirmation or equivalent recent authentication for password changes and account deletion. An Owner must transfer or delete every owned Family before deleting their account.
11. Rate-limit registration, login, verification resend, refresh, and password-reset commands using an in-process strategy appropriate for a single VM; no Redis or separate rate-limit service is needed.
12. Finish accounts when tests cover:
    1. Canonical email and username uniqueness, including concurrent registration.
    2. Username boundaries and rejection of uppercase, spaces, Unicode, and unsupported punctuation.
    3. Argon2id hashing, verification expiry/resend/single use, and login rejection before verification.
    4. JWT signature, algorithm, issuer, audience, expiry, and not-before validation.
    5. Refresh hashing, rotation, reuse detection, current logout, logout-all, and reset-driven revocation.
    6. Profile updates, sensitive-action authentication, account deletion constraints, and generic credential errors that do not disclose account existence.

## 3. Implement Families, hierarchy, and invitations

1. Represent ownership and membership with three entities:
   1. `families`: Snowflake `id`, `name`, non-null `owner_id`, `created_at`, `updated_at`, and optional `deleted_at` when recoverable deletion is retained. `owner_id` is the sole Owner relationship; there is no Owner membership row or Family revision column.
   2. `family_members`: Snowflake `id`, `family_id`, `user_id`, role (`manager` or `member`), `created_at`, `updated_at`, unique `(family_id, user_id)`, and an index beginning with `user_id`.
   3. `family_invitations`: Snowflake `id`, `family_id`, `invited_user_id`, `invited_by_user_id`, `expires_at`, `created_at`, and unique `(family_id, invited_user_id)`. Invitations are pending rows, contain no role/status/response timestamp, create a Member when accepted, and are deleted on acceptance, decline, cancellation, or expiry cleanup.
2. Enforce the role hierarchy in one server-side policy:

   | Action | Owner | Manager | Member |
   | --- | --- | --- | --- |
   | View and edit shared list | Yes | Yes | Yes |
   | Edit Family settings | Yes | Yes | No |
   | Invite or remove Members | Yes | Yes | No |
   | Promote/demote Managers | Yes | No | No |
   | Remove a Manager | Yes | No | No |
   | Transfer ownership | Yes | No | No |
   | Delete Family | Yes | No | No |

3. Limit each User to ten Families total, counting `families.owner_id` and `family_members`. Family creation and invitation acceptance lock the User row, count both relationships, and insert within the same transaction so two concurrent joins cannot exceed the limit.
4. Create a Family and establish its Owner atomically. Member-list responses synthesize the Owner entry from `families.owner_id`, then include Manager and Member rows.
5. Invite only by exact validated username or exact Snowflake User ID. The request supplies exactly one identifier:

   ```json
   {
     "username": "exact_username"
   }
   ```

   ```json
   {
     "user_id": "241895317264384000"
   }
   ```

   Resolve it privately and return `204 No Content`. Nonexistent, already participating, already invited, over-limit, and otherwise ineligible targets receive the same public `INVITEE_NOT_ELIGIBLE` error where distinguishing them would expose account existence. The API has no user search, suggestions, autocomplete, or general user lookup route.
6. Accepting an invitation rechecks its expiry and the ten-Family limit, creates a `member`, and deletes the invitation in one transaction. Promotion happens later through the member update route.
7. Transfer ownership with `{ "new_owner_id": "241895615438458880" }` in one transaction:
   1. Lock the Family and verify the caller is its Owner.
   2. Verify the new Owner is currently a Manager or Member and satisfies the Family limit.
   3. Delete the new Owner's membership row.
   4. Insert the old Owner as a Manager.
   5. Update `families.owner_id` and record the audit event.
8. Prevent an Owner from leaving or being removed before transfer. Every mutation must preserve the invariant that a User cannot simultaneously be a Family's `owner_id` and have a membership row in it.
9. Publish these routes:

   ```text
   GET    /v1/families
   POST   /v1/families
   GET    /v1/families/{family_id}
   PATCH  /v1/families/{family_id}
   DELETE /v1/families/{family_id}

   GET    /v1/families/{family_id}/members
   PATCH  /v1/families/{family_id}/members/{member_id}
   DELETE /v1/families/{family_id}/members/{member_id}
   POST   /v1/families/{family_id}/leave
   POST   /v1/families/{family_id}/transfer-ownership

   GET    /v1/invitations
   POST   /v1/invitations/{invitation_id}/accept
   POST   /v1/invitations/{invitation_id}/decline

   GET    /v1/families/{family_id}/invitations
   POST   /v1/families/{family_id}/invitations
   DELETE /v1/families/{family_id}/invitations/{invitation_id}
   ```

10. Store durable security/administrative events in `audit_events`: Snowflake `id`, nullable actor User ID, nullable Family ID, action code, target type and optional target ID, safe JSON metadata, request ID, and `created_at`. Record ownership, role, invitation, Family deletion, and account-security actions rather than every list keystroke.
11. Finish Families when real-MySQL integration tests prove:
    1. Exactly one Owner exists and appears in member responses without a membership row.
    2. Every role/action combination is enforced and cross-Family IDs never bypass scope.
    3. The tenth Family succeeds, the eleventh fails, and two concurrent joins from nine result in exactly ten.
    4. Exact username/ID invitations, duplicate prevention, generic ineligibility responses, expiry, acceptance, decline, and cancellation behave atomically.
    5. Ownership transfer succeeds completely or rolls back completely, including the old/new Owner membership changes.

## 4. Implement the current list, completion, and history

1. Use `shopping_list_items` as the Family's single current list; a separate `shopping_lists` table adds no useful state. Each item contains:
   1. Snowflake `id` and `family_id`.
   2. Required freeform `name VARCHAR(160)` so catalog coverage never blocks a purchase.
   3. Positive desired `quantity DECIMAL(12,3)` and an approved `unit`.
   4. Optional `note VARCHAR(500)`.
   5. `is_checked`.
   6. Nullable actual `purchased_quantity DECIMAL(12,3)` and nullable nonnegative `total_cents BIGINT` entered while shopping.
   7. Sparse sortable `position BIGINT`.
   8. `added_by_id`, `updated_by_id`, `created_at`, and `updated_at`.
2. Support add, edit, explicit check/uncheck, reorder, and delete. `added_by_id` never changes; edits update `updated_by_id`. Unchecking clears temporary purchase quantity and cents unless a later UI explicitly confirms retaining them.
3. Use last database commit wins for concurrent item changes. There is no item version or conditional-write header. An update racing with deletion returns `404 LIST_ITEM_NOT_FOUND`, and every query scopes the item by both `family_id` and `item_id`.
4. Treat a completed shopping trip as a historical result, not a persisted active workflow. There are no start, status, cancel, or active-trip fields/routes.
5. Store purchase history in:
   1. `shopping_trips`: Snowflake `id`, `family_id`, optional bounded freeform `supermarket_name`, `completed_by_id`, `completed_at`, and `created_at`. Create it only when at least one item was purchased.
   2. `shopping_trip_items`: Snowflake `id`, `trip_id`, non-null `source_item_id`, snapshotted name, requested quantity/unit, optional note, original `added_by_id`, actual purchased quantity, total cents, and `created_at`. `source_item_id` has no foreign key because the purchased source row is deleted while history remains immutable.
6. Complete the list with one request:

   ```json
   {
     "supermarket_name": "Market downtown",
     "default_unchecked_action": "keep",
     "unchecked_item_overrides": [
       {
         "item_id": "241896015117107200",
         "action": "discard"
       }
     ]
   }
   ```

   Derive the completing User from authentication. In one transaction, lock the Family and current items in a stable order, recheck authorization, and apply these outcomes:
   1. `purchased`: every checked item is copied into immutable trip history and removed from the current list.
   2. `keep`: an unchecked item stays in the current list, becomes unchecked, clears temporary purchase data, and is not written to history.
   3. `discard`: an unchecked item is removed and is not written to history.
   4. A per-item override wins over the default unchecked action.
   5. When nothing was purchased, apply keep/discard without creating an empty trip header.
7. Publish current-list and history routes:

   ```text
   GET    /v1/families/{family_id}/shopping-list
   POST   /v1/families/{family_id}/shopping-list/items
   PATCH  /v1/families/{family_id}/shopping-list/items/{item_id}
   DELETE /v1/families/{family_id}/shopping-list/items/{item_id}
   PUT    /v1/families/{family_id}/shopping-list/items/order
   POST   /v1/families/{family_id}/shopping-list/complete

   GET /v1/families/{family_id}/shopping-trips
   GET /v1/families/{family_id}/shopping-trips/{trip_id}
   ```

8. Return individual resources directly and named envelopes for collections, for example `{ "families": [] }`. Paginate growing history with opaque cursors; keep the small current list unpaginated while measurements support it. Authenticated responses use `Cache-Control: private, no-store`.
9. Finish shopping when tests prove:
   1. List creation, editing, state assignment, reordering, and deletion preserve attribution and Family scope.
   2. Quantities never pass through binary floating point and all money remains integer cents.
   3. Purchased items leave the list and appear once in immutable history with the original adder and completing User.
   4. Kept and discarded items never enter history, and a zero-purchase completion creates no header.
   5. Completion is atomic; concurrent completion or list mutation cannot create duplicate history or partial loss.

## 5. Build the React Native application

1. Use Expo, React Native, Expo Router, strict TypeScript, TanStack Query, the generated API client, and operating-system secure storage. Keep native screen/router code separate from the future DOM client while sharing platform-neutral query keys, money/quantity helpers, validation presentation, translations, and design tokens where useful.
2. Implement authentication screens for login, registration, verification/resend, forgotten password, and reset password. Store access and refresh credentials only in secure storage, refresh the 30-minute JWT shortly before or after expiry, and clear credentials when refresh is expired or revoked.
3. Implement authenticated screens for:
   1. Family chooser and creation.
   2. Received invitations.
   3. Current Family grocery list.
   4. List completion with keep/discard decisions.
   5. Purchase-history list and detail.
   6. Family members, invitations, and settings according to the current role.
   7. Account profile, password change, and current/all-session logout.
4. Cache the last successful Family/list snapshot locally with its Family ID and `last_checked_at`. Render it immediately on launch while refreshing from the API.
5. When offline or the API is unreachable, continue showing the cached snapshot and display `Offline · last checked <local time>` at the upper-left edge of the grocery-list button in the bottom navigation/footer. Make stale state visually clear and disable/fail mutations before presenting them as successful.
6. Do not queue offline mutations. Refresh on app launch, foreground, reconnect, pull-to-refresh, and every successful mutation. Clear secure credentials and cached private Family data on logout.
7. Do not request notification permission or create notification settings.
8. Finish mobile when automated/component tests and supported-device journeys cover registration through verification, Family creation/invitation acceptance, list editing, purchase completion, history, ownership transfer, logout, airplane-mode launch, process death, restart, reconnect, cache timestamps, offline warning accessibility, and cache cleanup.

## 6. Provision and deploy the production system

1. Keep OCI usage within Always Free: one `MySQL.Free` DB system and only `VM.Standard.E2.1.Micro` AMD application instances. The initial deployment needs one VM, but the network rules must support another authorized VM without exposing MySQL.
2. Bootstrap an OCI Object Storage bucket under `infra/terraform/bootstrap`, then configure the production root with the native backend:

   ```hcl
   terraform {
     backend "oci" {}
   }
   ```

   Commit a placeholder-only `backend.oci.tfbackend.example`, ignore the real `backend.production.tfbackend`, enable bucket versioning for Terraform-state recovery, restrict its IAM permissions, and initialize with:

   ```text
   terraform init -backend-config=backend.production.tfbackend
   ```

   Bucket, namespace, region, state key, and a non-secret OCI profile name belong in backend configuration; credentials remain in OCI configuration, environment variables, or CI secrets. Use the OCI backend's state locking.
3. Build one VCN with an initial layout such as:
   1. VCN: `10.0.0.0/16`.
   2. Public VM subnet: `10.0.1.0/24` with an Internet Gateway.
   3. Private MySQL subnet: `10.0.2.0/24` with no public route or endpoint.
   4. No NAT Gateway, since it risks cost and is unnecessary for the initial topology.
4. Query `oci_identity_availability_domains` in the production root and choose `availability_domain_index` from `terraform.tfvars`. Validate the index, support single-AD regions, and allow the operator to change it if E2 Micro capacity is unavailable rather than hard-coding a tenancy-specific AD name.
5. Commit `terraform.tfvars.example` and ignore the real tfvars. Include non-secret deployment choices such as:

   ```hcl
   availability_domain_index = 0
   admin_ipv4_cidr            = "203.0.113.10/32"
   compute_shape              = "VM.Standard.E2.1.Micro"
   cloudflare_ipv4_cidrs      = [
     # Reviewed values from https://www.cloudflare.com/ips-v4/
   ]
   ```

6. Declare `cloudflare_ipv4_cidrs` as a non-secret `set(string)` Terraform variable. Pin and review the complete list published for connections to proxied IPv4 origins instead of downloading mutable security rules during `terraform plan`. When Cloudflare announces a change, add new CIDRs and verify traffic before removing deprecated ones. Add an IPv6 variable and rules only if the VCN, VM, and origin DNS later expose IPv6.
7. Give each application VM a stable private IPv4, public IPv4, SSH public key, compute NSG, stable Snowflake node ID, and the MySQL-client NSG when authorized. Configure stateful ingress as follows:
   1. TCP 22 from `admin_ipv4_cidr` only.
   2. TCP 443 from every `cloudflare_ipv4_cidrs` entry only.
   3. No public TCP 80, 8080, 8443, or 3306.
   4. No administrator bypass to the HTTPS origin unless a temporary, reviewed troubleshooting rule is deliberately applied.
8. Give the HeatWave DB system only a private endpoint. Its NSG permits stateful TCP 3306 from the MySQL-client NSG, not from a public CIDR or the full VM subnet. Additional VMs gain access by attaching that client NSG. Require TLS and a least-privilege API database user.
9. Keep Cloudflare provisioning outside Terraform and this implementation plan. The external edge prerequisite is a proxied `api.example.com` record targeting the VM's public IPv4, Full (strict) origin encryption, authenticated-response cache bypass, and any desired edge rate limit. There is no Cloudflare provider, Terraform-managed zone, OCI Load Balancer, or Cloudflare Tunnel in this architecture.
10. Use separate certificates for the two TLS hops:
    1. Cloudflare manages its browser-facing Universal SSL certificate.
    2. Echo presents a Cloudflare Origin CA certificate covering `api.example.com` or the approved wildcard.
    3. Create the Origin CA certificate outside Terraform so its private key never enters Terraform configuration, plans, or state.
    4. Store the private key in Ansible Vault or inject it from the deployment secret store; never commit its plaintext or bake it into the container image.
11. Make Ansible install the origin material on the VM:

    ```text
    /etc/ostos/tls/origin.crt
    /etc/ostos/tls/origin.key
    ```

    Give the directory and private key only the permissions required by the fixed non-root UID/GID used by the API container. Mark certificate-copy tasks `no_log`, notify the API restart handler when either file changes, and keep the files after container replacement.
12. Run Echo TLS inside the container rather than adding a reverse proxy. Mount the two host files read-only at `/run/secrets/ostos/origin.crt` and `/run/secrets/ostos/origin.key`, listen on unprivileged container port `8443`, and publish only host `443:8443`. Configure Echo v5 through `StartConfig.StartTLS`, with the paths supplied as runtime configuration:

    ```go
    start := echo.StartConfig{
        Address:         ":8443",
        HideBanner:      true,
        GracefulTimeout: 10 * time.Second,
        TLSConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    }

    err := start.StartTLS(ctx, server, cfg.TLSCertFile, cfg.TLSKeyFile)
    ```

    Treat any error other than the expected graceful-shutdown result as fatal. Fail startup when either file is absent, unreadable, malformed, expired, or does not contain a matching key pair.
13. Configure the VM with Ansible:
    1. Use a supported minimal Linux image and a locked-down `ostos` service identity with a fixed numeric UID/GID shared by the Compose definition.
    2. Install a pinned Docker Engine and Compose plugin, then run only the API as a long-lived application container.
    3. Align the host firewall with the OCI NSG and enforce the same Cloudflare-only TCP 443 policy in the `DOCKER-USER` chain so a published Docker port does not bypass host rules.
    4. Store runtime configuration in root-owned restricted files and the TLS material under `/etc/ostos/tls`.
    5. Upload a versioned, compressed `docker save` archive produced by CI, load it locally, and select its immutable image tag in `compose.production.yml`; a registry is unnecessary for this deployment size.
    6. Configure Compose restart behavior, health checking, resource limits suitable for E2 Micro, read-only mounts, and the minimum required writable temporary paths.
    7. Send container logs to bounded journald and retain the previous image tag/archive for rollback.
    8. Verify Ansible idempotence with a second run.
14. Provision SES independently of Cloudflare automation. Restrict AWS credentials to sending approved account emails from the verified identity; publish the required SES verification, DKIM, SPF, and DMARC records through the externally managed DNS process. Keep AWS, database, origin-key, and other secrets outside Terraform source, tfvars, state outputs, images, and logs.
15. Use OCI Object Storage for future application objects. Do not provision a product-media bucket before product media exists. GCP's regional Always Free storage may remain a documented alternative, not a second active dependency.
16. Implement lightweight operations with structured `slog` JSON containing request ID, route template, status, duration, safe User ID, and stable error code. Use Docker health status, systemd/journald, and OCI instance metrics. Add alerts only for VM unreachability, disk pressure, repeated container restart, and health failure; a Prometheus stack is unjustified here.
17. Publish one operational endpoint outside `/v1`:

    ```text
    GET /health
    ```

    It returns `200` when the process and required database are ready and `503` when the process responds but MySQL is unavailable, without leaking topology or credentials. Separate liveness/readiness routes are unnecessary for this single-container deployment.
18. Apply baseline security controls:
    1. Exact CORS origins and bounded request bodies.
    2. Parameterized, Family-scoped GORM queries and real `sql.Tx` transactions with consistent lock ordering.
    3. Context deadlines for HTTP, MySQL, SES, and provider calls and a small E2-Micro-appropriate connection pool.
    4. Explicit trusted-proxy handling; accept Cloudflare client-IP headers only when the immediate peer is within the same pinned Cloudflare CIDRs admitted by the NSG.
    5. No logging of passwords, JWTs, bearer tokens, invite identifiers, credentials, Terraform state, or private shopping content.
    6. No continuous location collection.
19. Do not add application backup jobs. Document that this deployment has no project-managed recovery guarantee; leave mandatory/no-cost provider protections enabled without describing them as a tested recovery design.
20. Run delivery gates in CI:
    1. Backend: `gofmt`, `go vet`, static analysis, dependency vulnerability scanning, unit tests, race tests where compatible, and real-MySQL integration tests.
    2. Migrations: build the independent module, migrate an empty MySQL database, inspect constraints/indexes, and execute GORM compatibility tests.
    3. API: lint OpenAPI, detect breaking changes, regenerate types/clients, and fail on uncommitted generated output.
    4. Mobile: format/lint, strict TypeScript, unit/component tests, build, and critical-device flows.
    5. Container: build the Linux AMD64 image, scan it, run it as the production non-root UID, verify read-only TLS mounts, and test graceful shutdown and restart behavior.
    6. Infrastructure: `terraform fmt -check`, `terraform validate`, reviewed saved plans, E2-only/public-database policy checks, Cloudflare-CIDR-only HTTPS checks, Ansible syntax/lint, and convergence testing.
21. Deploy in this order: reviewed Terraform changes, forward-compatible migrations, CI-built versioned image archive, Ansible host/certificate configuration, `docker load`, Compose image-tag update, container replacement, then health and external smoke tests. A failed application deployment restores the previous image tag without attempting a destructive schema downgrade.
22. Finish production deployment when tests prove restricted `/32` SSH, TCP 443 access from every pinned Cloudflare origin CIDR and no other source, direct Full (strict) TLS to Echo, private TLS MySQL access from approved NSGs only, backend locking, E2 Micro container viability, Ansible idempotence, and reproducible deployment from an empty VM.

## 7. Add the complete authenticated website

1. Start this near-future phase only after the API and mobile critical path are stable. Use React, Vite, TanStack Router with route-level code splitting, TanStack Query, strict TypeScript, and the generated API client. Do not add Next.js, SSR, server actions, a JavaScript BFF, or duplicated authorization logic.
2. Implement the full current product rather than a landing page:

   ```text
   /
   /login
   /register
   /verify-email
   /forgot-password
   /reset-password

   /app
   /app/families
   /app/families/new
   /app/families/{family_id}/list
   /app/families/{family_id}/history
   /app/families/{family_id}/history/{trip_id}
   /app/families/{family_id}/members
   /app/families/{family_id}/invitations
   /app/families/{family_id}/settings
   /app/account
   ```

3. Make `/` redirect authenticated users to `/app` and unauthenticated users to `/login`. `/app` redirects to the last active Family list or Family chooser.
4. Deploy the static build to Cloudflare Pages and configure SPA fallbacks so direct navigation and browser refresh work for every dynamic route.
5. Do not create `/features`, `/terms`, `/privacy`, or `/support` as product routes. Revisit only a required privacy-policy URL before public app-store distribution.
6. Finish the website when Playwright covers authentication, every current account/Family/list/completion/history journey, role-dependent controls, direct dynamic-route navigation, refresh, and production build output.

## 8. Extend the product catalog and price comparison later

1. Begin this work only after real usage proves the core list and history stable. A complete supermarket inventory is not required for grocery-list usefulness and should not delay the current scope.
2. Model future concepts separately:
   1. Category, such as soap or soda.
   2. Brand, such as Dove or Coca-Cola.
   3. Manufacturer/brand-owning organization.
   4. Product family.
   5. Exact sellable variant containing package size and optional GTIN/barcode.
   6. Retailer-specific listing.
   7. Physical supermarket branch.
   8. Time-stamped branch assortment assertion.
   9. Time-bounded price or promotion.
3. Keep the list item's freeform name after catalog introduction. Link it to an exact variant only when the shopper deliberately selects or scans one; a generic entry such as `soap` remains valid. Brand belongs to the product relationship and bottle/package size belongs to the exact variant.
4. Store licensed future images in OCI Object Storage under immutable keys with provenance, content hash, dimensions, and moderation metadata. Never hotlink or scrape retailer images without rights.
5. Integrate one retailer provider first behind a provider adapter. Match exact validated GTIN, then reviewed provider mappings, then human-reviewed brand/name/package candidates; never merge products automatically because their titles look similar.
6. Treat absent retailer data as unknown rather than out of stock. Every comparison must show its source, branch, retrieval time, validity period, promotion conditions, and freshness.
7. Add neither a search cluster nor the HeatWave analytical accelerator until measured queries justify them. Public product/price pages and SSR become separate decisions only when catalog quality and acquisition value exist.
