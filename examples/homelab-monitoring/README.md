# Homelab Monitoring Stack (Prometheus + Grafana + Loki) with Terraform

Full observability stack for your homelab or single-node deployment, defined in Terraform via the `xRizur/dockercompose` provider. Three services, configured to discover each other automatically — Prometheus scrapes itself + Grafana + Loki, and Grafana can immediately add Prometheus and Loki as data sources.

**What you get:**

- **Prometheus 2.54** at `http://localhost:9090` — metrics scraping with a starter scrape config (`prometheus.yml`)
- **Loki 3.2** at `http://localhost:3100` — log aggregation with TSDB filesystem storage
- **Grafana 11.3** at `http://localhost:3000` — dashboards (`admin` / configurable password)
- All three on a private bridge network with healthchecks
- Persistent named volumes for metrics, logs, and Grafana data
- Variables for ports + admin password

## Verified

✅ Tested end-to-end on `2026-04-26` (Terraform 1.14.3, Docker 29.1.5). All three services reach `healthy` status; `/-/healthy` (Prometheus), `/ready` (Loki), and `/api/health` (Grafana) all respond `200 OK` after ~30-45s.

> **Note on Loki paths:** the Loki container runs as a non-root user (`uid 10001`). The bundled `loki-config.yml` writes data under `/loki` (not `/tmp/loki`) because that's the only path with the right permissions on a freshly-created named volume. If you change the path, you'll also need to chown the volume or run Loki as root — both are nastier than just keeping `/loki`.

## Run

```bash
terraform init
terraform apply
# Wait ~30s for healthchecks to settle, then open http://localhost:3000
terraform destroy
```

Initial Grafana login: `admin` / `admin` (change via `-var="grafana_admin_password=..."`).

## Add data sources in Grafana

After login, go to **Connections → Data sources**:

| Type | URL |
|---|---|
| Prometheus | `http://prometheus:9090` |
| Loki | `http://loki:3100` |

(Use the in-network DNS names, not `localhost` — Grafana resolves siblings on the `monitoring` network.)

## Customize

```bash
terraform apply \
  -var="grafana_port=3001" \
  -var="grafana_admin_password=$(openssl rand -hex 16)"
```

Add scrape targets by editing `prometheus.yml` (re-apply to push the change to the running Prometheus container).

## Architecture

```
+------------+   +------------+   +-----------+
| Prometheus |   |    Loki    |   |  Grafana  | ← :3000
|   :9090    |   |   :3100    |   |  :3000    |
+------+-----+   +-----+------+   +-----+-----+
       |               |                |
       |               +-------+--------+
       +-----------------------+
              monitoring (bridge)
```

## Extending it

- **Node Exporter / cAdvisor** — add another service block for host or container metrics; add the target to `prometheus.yml`.
- **Alertmanager** — add `alertmanager` service, mount its config, and set `alerting:` block in Prometheus config.
- **Promtail** — add Promtail as a sidecar to ship logs to Loki.

## Keywords

terraform prometheus grafana loki, homelab monitoring as code, terraform observability stack, docker compose terraform monitoring, self-hosted prometheus grafana terraform, dockercompose_stack monitoring
