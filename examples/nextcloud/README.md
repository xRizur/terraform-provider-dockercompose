# Nextcloud + Postgres + Redis with Terraform

Self-hosted Nextcloud deployment as Terraform code. Three services (Nextcloud app + Postgres database + Redis cache), defined entirely in HCL, with healthcheck-aware startup so Nextcloud only boots once Postgres is ready.

**What you get:**

- **Nextcloud 30 (Apache)** at `http://localhost:8081`
- **Postgres 16 Alpine** for the metadata DB (with `pg_isready` healthcheck)
- **Redis 7 Alpine** for file locking + caching (significant speedup vs. the default file-based locking)
- Auto-installed admin user (no install-wizard click-through needed)
- Three named volumes: app code, user data, database
- Variables for port, credentials, trusted domains

## Verified

✅ Tested end-to-end on `2026-04-26` (Terraform 1.14.3, Docker 29.1.5). All three services reach `running`; root URL responds `HTTP 302` (redirect to the login page) once auto-install completes (~1-2 min on first apply while Nextcloud initializes).

## Run

```bash
terraform init
terraform apply
# First boot takes 1-2 min while Nextcloud auto-installs.
# Then open http://localhost:8081 and log in.
terraform destroy
```

Default credentials: `admin` / `change-me-admin` (override via `-var=...`).

## Customize

```bash
terraform apply \
  -var="nextcloud_port=8443" \
  -var="admin_password=$(openssl rand -hex 24)" \
  -var="db_password=$(openssl rand -hex 24)" \
  -var='trusted_domains=cloud.example.com localhost'
```

## Architecture

```
        +----------------------+
        | Nextcloud (Apache)   | ← :8081
        |  /var/www/html       | ← nc_app_data
        |  /var/www/html/data  | ← nc_user_data
        +----+-----------+-----+
             |           |
             v           v
   +-----------+   +----------+
   | Postgres  |   |  Redis   |
   | nc_db_data|   |  (mem)   |
   +-----------+   +----------+
              nc_net (bridge)
```

## Production checklist

This example optimizes for "works locally in 2 minutes." Before you put it in production:

- [ ] Move secrets out of tfvars into Vault / SOPS / cloud secret manager
- [ ] Add Traefik or Caddy in front for TLS termination + Let's Encrypt
- [ ] Set `OVERWRITEPROTOCOL=https` and add your real domain to `trusted_domains`
- [ ] Configure backups for `nc_user_data` and `nc_db_data` (volumes are persistent but not backed up)
- [ ] Increase `resource_limits_memory` based on user count
- [ ] Consider separating user data onto a host-mount or NFS volume for easier backup

## Why no Traefik in this example?

Traefik adds DNS, certificate, and ACME complexity that breaks "clone and run." This example stays minimal so you can verify the core Nextcloud + Postgres + Redis combo works, then layer reverse proxy / TLS on top in your own variant. See the homelab-monitoring example for a similar private-network pattern you can extend.

## Keywords

terraform nextcloud, self-hosted nextcloud IaC, docker compose nextcloud terraform, nextcloud postgres redis terraform, homelab nextcloud, dockercompose_stack nextcloud example
