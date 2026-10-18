# Production VM recovery and health response

## Lost or reclaimed VM

1. Preserve Terraform state and inspect OCI events before changing anything.
2. Run a reviewed Terraform plan. Recreate only the missing compute resource;
   do not replace the protected MySQL DB system or state bucket.
3. Update the proxied Cloudflare A record if the public IP changed.
4. Update the Ansible inventory, rerun the production playbook with the current
   image archive and Vault, then run all network/TLS/database smoke checks.
5. Assign a Snowflake node ID that cannot overlap a concurrently running API
   process. A replacement for a definitively dead VM may reuse its ID.

## Health alerts

OCI Monitoring sends infrastructure-unreachable alarms through the configured
Notifications topic after its email subscription is confirmed. The
`ostos-local-health.timer` checks root disk pressure, container absence/health,
and repeated restarts every five minutes and writes priority-2 structured
journald events with `OSTOS_ALERT`.

Inspect failures with:

```sh
systemctl status ostos-local-health.service
journalctl -u ostos-local-health.service -u docker.service --since=-1h
docker inspect ostos-api
docker logs --since=30m ostos-api
df -h /
```

Never paste runtime environment files, Terraform state, bearer values, email
tokens, or private shopping content into an incident ticket. If credentials are
suspected, rotate the affected JWT/email encryption/AWS/DB secret, redeploy, and
revoke the old credential at its provider. Existing access JWTs can remain valid
for at most 30 minutes; database-backed refresh sessions can be revoked.

Database loss is not recoverable through this project: no application backup
job or tested restore path exists. Stop and reassess the architecture/budget
instead of claiming an RPO or RTO that the system does not provide.
