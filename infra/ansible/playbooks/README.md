# Ansible playbooks

`production.yml` converges an empty Oracle Linux 9 VM and deploys a CI-built
AMD64 image archive. It deliberately does not run Terraform or migrations.

```sh
ansible-galaxy collection install -r requirements.yml
ansible-playbook --syntax-check playbooks/production.yml
ansible-playbook playbooks/production.yml --ask-vault-pass
ansible-playbook playbooks/production.yml --ask-vault-pass --check
```

The final check-mode run must report no unintended changes.
