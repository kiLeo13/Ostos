# Production deployment and rollback

## Prepare external services

1. Create a Cloudflare Origin CA certificate for the exact API hostname and set
   the zone to Full (strict). Store the certificate and key only in the encrypted
   Ansible Vault file.
2. Create a proxied `api` A record pointing to Terraform's `api_public_ip`.
   Bypass caching for authenticated/private API responses.
3. Verify the AWS SES identity, DKIM, SPF, and DMARC records, leave the SES
   sandbox when required, and create send-only credentials.
4. Create the MySQL schema and least-privilege migration/application users over
   the VM's private path. The API account receives no schema privileges.

## Prepare Ansible

Copy `inventories/production.example.yml` to the ignored `production.yml`.
Replace host values and the checked-in non-secret production variables. Copy
`vault.example.yml` to `vault.yml`, replace every value, and encrypt it:

```sh
cd infra/ansible
ansible-vault encrypt inventories/group_vars/production/vault.yml
ansible-galaxy collection install -r requirements.yml
ansible-playbook --syntax-check playbooks/production.yml
```

## Build and deploy

Build on AMD64 CI, never on the E2 Micro VM:

```sh
docker build --platform linux/amd64 -t ostos-api:${VERSION} backend
docker save ostos-api:${VERSION} | gzip > infra/ansible/artifacts/ostos-api-${VERSION}.tar.gz
```

Update `ostos_api_version`, `ostos_api_image`, and `ostos_api_archive_local`.
Run forward-compatible migrations first, then converge the host:

```sh
ansible-playbook playbooks/production.yml --ask-vault-pass
ansible-playbook playbooks/production.yml --ask-vault-pass --check
```

The second run must report no unintended changes. Verify `docker inspect
ostos-api`, `systemctl status ostos-local-health.timer`, bounded journald use,
and `https://api.example.com/health` through Cloudflare. Exercise login, token
refresh, Family listing, one list mutation, and history read without logging any
payload or credential.

## Roll back application code

Do not downgrade the schema. Select the previous retained image tag in the
Ansible variables and its retained archive, then rerun the playbook. Verify the
health endpoint and smoke flow. If the previous binary is not compatible with
the expanded schema, repair forward with a new image instead.

## Rotate the Origin CA certificate

Create a replacement that overlaps the current certificate. Replace both Vault
values, rerun Ansible, and verify Full (strict) before revoking the former
certificate. The certificate-copy tasks suppress secret output and restart only
the API container when material changes.
