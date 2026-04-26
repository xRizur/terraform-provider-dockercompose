# Example: `dockercompose_stack` — full HCL Compose stack

A complete monitoring stack defined entirely in HCL — no `docker-compose.yml` file involved. Demonstrates the most common building blocks of `dockercompose_stack`:

- Two services (`database`, `grafana`) with `depends_on` ordering
- Named volumes (`db_data`, `grafana_data`) declared as top-level `volume` blocks
- A private network (`monitoring_net`) declared as a top-level `network` block
- Healthcheck on the Postgres service so Grafana waits for the DB to be ready
- CPU / memory resource limits
- `remove_volumes_on_destroy = true` so `terraform destroy` cleans up persistent data

## Run

```bash
terraform init
terraform apply
```

Once it's up:

```bash
docker compose -p metrics-stack ps
# Grafana: http://localhost:3000 (admin/admin)
# Postgres: localhost:5432 (admin / secretpassword)
```

Tear it down:

```bash
terraform destroy
```

## What to change

- Adjust `image` tags to pin specific versions instead of `latest`.
- Replace the inline `POSTGRES_PASSWORD` with a Terraform variable or secret store before using anywhere real.
- Add more services by appending another `service { ... }` block — use `depends_on = ["..."]` to control startup order.
- For remote deployment, configure `provider "dockercompose" { host = "ssh://user@server" }`.

## Related

- [`dockercompose_project` example](../dockercompose_project/) — same idea but reading from an existing `docker-compose.yml`
- [Full attribute reference](https://registry.terraform.io/providers/xRizur/dockercompose/latest/docs/resources/stack)
