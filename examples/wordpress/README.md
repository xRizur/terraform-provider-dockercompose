# WordPress + MySQL with Terraform (Docker Compose)

Spin up a complete WordPress site with MySQL backend in one `terraform apply`. Two services, two named volumes, healthcheck on the database so WordPress only starts once MySQL is ready. Useful as a starter template for self-hosted WordPress, dev environments, or staging instances.

**What you get:**

- WordPress 6 (PHP 8.2 / Apache) on `http://localhost:8080`
- MySQL 8.0 with persistent storage in a named volume
- Healthcheck-aware startup ordering (`depends_on` + `healthcheck_test`)
- Variables for ports and credentials
- Per-service runtime status as Terraform outputs

## Verified

✅ Tested end-to-end on `2026-04-26` (Terraform 1.14.3, Docker 29.1.5, provider `xRizur/dockercompose` 1.1.x). `terraform apply` brings both containers to `running`, WordPress responds with `HTTP 302` redirect to the install wizard. `terraform destroy` cleans both volumes when `remove_volumes_on_destroy = true`.

## Run

```bash
terraform init
terraform apply
# Open http://localhost:8080 and complete the install wizard
terraform destroy
```

## Customize

```bash
terraform apply \
  -var="wordpress_port=8888" \
  -var="db_root_password=$(openssl rand -hex 16)" \
  -var="db_password=$(openssl rand -hex 16)"
```

For production, move secrets to a tfvars file ignored by git, or pull them from Vault / AWS Secrets Manager / SOPS.

## Architecture

```
        +---------------------+
        |  wordpress (Apache) | ← :8080
        |  /var/www/html      | ← wp_files volume
        +----------+----------+
                   | depends_on
                   v
        +---------------------+
        |    db (MySQL 8)     |
        |  /var/lib/mysql     | ← wp_db_data volume
        +---------------------+
                wp_net (bridge)
```

## Common tweaks

- **Persistent uploads only** (no full WP files in volume): mount `wp_files:/var/www/html/wp-content` instead of `:/var/www/html`.
- **TLS / domain**: add a Traefik or Caddy service in front and set `WORDPRESS_CONFIG_EXTRA` for HTTPS detection.
- **Pin WP version**: change `wordpress:6-php8.2-apache` to `wordpress:6.7.2-php8.2-apache`.
- **Move to remote host**: set `provider "dockercompose" { host = "ssh://user@server" }` — same config works against any Docker daemon.

## Keywords

terraform wordpress, terraform docker compose wordpress, self-hosted wordpress IaC, docker compose terraform provider, wordpress mysql terraform, dockercompose_stack wordpress example
