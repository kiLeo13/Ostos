# Production environment

This root provisions the single Always Free OCI production topology. It does
not manage Cloudflare, DNS, Origin CA keys, SES, or application deployment.

1. Apply `../../bootstrap` and copy its outputs into
   `backend.production.tfbackend` from the checked-in example.
2. Copy `terraform.tfvars.example` to the ignored `terraform.tfvars` and replace
   every placeholder. Keep the MySQL password out of that file.
3. Export `TF_VAR_mysql_admin_password` through the shell or CI secret store.
4. Run `terraform init -backend-config=backend.production.tfbackend`.
5. Run `terraform plan -out=production.tfplan`, review it, then apply that exact
   saved plan.

The initial run may fail for lack of E2 Micro capacity. Change only
`availability_domain_index` to another returned AD and re-plan. Never change the
shape to escape the free-tier guard.
