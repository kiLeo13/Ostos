# Production provisioning

Status: automation implemented; an operator must perform and record the live OCI and Cloudflare verification.

## Prerequisites

1. Confirm the OCI home region currently offers both `VM.Standard.E2.1.Micro`
   capacity and the `MySQL.Free` shape.
2. Install Terraform in the supported range and configure a non-secret OCI
   profile name whose credentials remain in the OCI configuration or CI store.
3. Create an OCI IAM group for state operators and an SSH key dedicated to this
   VM. Never put the private key in this repository.
4. Choose the production compartment and an alert email address.

## Bootstrap remote state

```sh
cd infra/terraform/bootstrap
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform plan -out=bootstrap.tfplan
terraform apply bootstrap.tfplan
terraform output
```

Store the bootstrap state in encrypted operator storage. Copy the bucket and
namespace outputs into
`infra/terraform/environments/production/backend.production.tfbackend` from its
example. OCI credentials do not belong in that file.

## Provision production

Copy `terraform.tfvars.example` to the ignored `terraform.tfvars`, replace all
placeholder OCIDs, `/32`, SSH public-key path, and alarm email, then export the
MySQL administrator password through `TF_VAR_mysql_admin_password`.

```sh
cd infra/terraform/environments/production
terraform init -backend-config=backend.production.tfbackend
terraform fmt -check -recursive ../..
terraform validate
terraform plan -out=production.tfplan
terraform show production.tfplan
terraform apply production.tfplan
terraform output
```

Confirm the OCI Notifications email subscription. Terraform assigns a
destruction-protected reserved public IPv4 to the VM, so normal instance
replacement can reattach the same address. If E2 capacity is absent,
change only `availability_domain_index` and re-plan. Do not substitute a paid
shape. Preserve the resulting public IP, private MySQL hostname, and Snowflake
node ID for inventory/runtime configuration.

## Required network verification

From the administrator `/32`, prove SSH succeeds. From any other source, prove
SSH fails. After Cloudflare is configured, prove TCP 443 succeeds through every
currently published Cloudflare origin range and direct non-Cloudflare access
fails. Confirm there are no public listeners/rules for 80, 8080, 8443, or 3306.

From the VM, connect to the Terraform output `mysql_hostname` using `tls=true`
and the least-privilege application account. Confirm a machine without the
MySQL-client NSG cannot connect. Store evidence with the reviewed plan; do not
store credentials or state.

There is no project-managed database backup or recovery guarantee. The DB and
state bucket are destruction-protected in Terraform, and state bucket versioning
is enabled, but neither fact is a substitute for a tested data restore.
