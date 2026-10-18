# Terraform state bootstrap

This deliberately small stack uses local state once to create a private,
versioned OCI Object Storage bucket and a bucket-scoped IAM policy. Its local
state contains resource identifiers and must be moved to encrypted operator
storage after the production backend is initialized; it must never be committed.

```sh
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform plan -out=bootstrap.tfplan
terraform apply bootstrap.tfplan
terraform output
```

Create the named IAM group before applying and add only the human/CI identities
that operate production Terraform. The bootstrap identity itself needs authority
to create the bucket and policy. Do not put OCI API keys in tfvars.
