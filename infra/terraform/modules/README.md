# Terraform modules

The production root composes three deliberately narrow OCI modules:

- `network`: VCN, public/private subnets, empty security lists, and NSGs.
- `compute`: one guarded E2 Micro instance with stable addressing.
- `database`: one deletion-protected private `MySQL.Free` DB system.

Cloudflare is intentionally not a Terraform provider dependency.
