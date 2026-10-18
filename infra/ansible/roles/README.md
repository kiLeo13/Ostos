# Ansible roles

- `common` creates the fixed service identity and bounded journald policy.
- `docker` installs pinned Docker Engine and Compose packages.
- `firewall` mirrors the OCI NSG at firewalld and `DOCKER-USER` layers.
- `api` installs Origin CA material, runtime configuration, and a versioned API
  image archive.

There are intentionally no tunnel, proxy, metrics, or backup roles.
