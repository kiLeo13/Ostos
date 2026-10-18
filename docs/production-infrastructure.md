# Production infrastructure implementation

Status: step 6 source automation implemented and locally validated on 2026-10-18; no OCI resources were applied from this workspace.

## Deployed topology

The production root creates one `10.0.0.0/16` VCN, a public
`10.0.1.0/24` compute subnet, and a route-less private `10.0.2.0/24`
MySQL subnet. Both subnets use empty security lists so Network Security Groups
are the only ingress policy. The one E2 Micro VM receives stable private address
`10.0.1.10`, a destruction-protected reserved public IPv4, the compute NSG, and
the MySQL-client NSG. The private `MySQL.Free` DB system receives address
`10.0.2.10` and accepts stateful TCP 3306 only from the MySQL-client NSG.

The compute NSG permits TCP 22 from the configured administrator `/32` and TCP
443 from the checked-in, reviewed Cloudflare IPv4 set. It does not admit 80,
8080, 8443, or 3306. The host repeats these controls with firewalld and a
`DOCKER-USER` chain. Compose binds IPv4 host 443 to non-root container port 8443;
Echo terminates Cloudflare Origin CA TLS without a reverse proxy.

## State and secret boundaries

The bootstrap root creates the versioned, private state bucket and a policy
scoped to a supplied IAM group OCID. The production root uses Terraform's native
OCI backend and locking. The MySQL administrator password is supplied through
`TF_VAR_mysql_admin_password`; OCI necessarily persists that resource argument
inside restricted Terraform state, but it is absent from source, tfvars, plans
committed to Git, and outputs.

Ansible Vault supplies the application DSN, application cryptographic secrets,
AWS SES credentials, and Origin CA material. Ansible writes them only to
root-owned files, suppresses certificate/runtime task output, and mounts TLS
files read-only. CI supplies a versioned AMD64 image archive; the VM never builds
the application and retains only the current and previous archives.

## Operations

OCI Notifications receives the infrastructure-unreachable alarm after its email
subscription is confirmed. A five-minute systemd timer emits priority-2
structured journald events for root disk pressure, absent/unhealthy containers,
and three or more container restarts. Journald retention is bounded. There is no
Prometheus stack, load balancer, tunnel, reverse proxy, registry, or application
backup job.

## Local verification record

- Terraform `fmt -check`, bootstrap/production `validate`, and three native
  policy tests pass with OCI provider 8.29.0.
- Policy tests prove the expected free shapes and reject a paid compute shape
  and non-`/32` SSH ingress.
- Ansible syntax and the `production` ansible-lint profile pass.
- Production Compose parses, and the Linux AMD64 image builds with fixed
  UID/GID 65532 and the binary health check.
- Backend tests/vet, OpenAPI lint/regeneration/type checking, migrations tests,
  and all 28 mobile tests pass.

Live NSG/firewall reachability, Full (strict) TLS, private HeatWave TLS,
Terraform backend locking, Ansible second-run idempotence, SES delivery, OCI
alarm delivery, and E2 resource viability remain operator acceptance checks.
