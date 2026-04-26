# Examples

Runnable examples for the [`xRizur/dockercompose`](https://registry.terraform.io/providers/xRizur/dockercompose/latest) Terraform provider. Each subfolder is a standalone example — `cd` in, `terraform init`, `terraform apply`.

## End-to-end use cases

These are real-world stacks you can copy-paste, run, and use today. Each one was tested against a live Docker daemon (where applicable — see per-example README for verification status).

| Example | Stack | Tested |
|---|---|---|
| [`wordpress/`](wordpress/) | WordPress 6 + MySQL 8 | ✅ apply / HTTP probe / destroy |
| [`homelab-monitoring/`](homelab-monitoring/) | Prometheus + Loki + Grafana | ✅ apply / 3× health probe / destroy |
| [`nextcloud/`](nextcloud/) | Nextcloud 30 + Postgres 16 + Redis 7 | ✅ apply / HTTP probe / destroy |
| [`remote-vps/`](remote-vps/) | nginx deployed to a remote VPS over SSH

## Reference snippets

Smaller, single-feature examples used for the Terraform Registry docs:

- [`provider/`](provider/) — minimal provider block (local socket and remote-host options)
- [`resources/dockercompose_stack/`](resources/dockercompose_stack/) — full HCL stack template (Postgres + Grafana)
- [`resources/dockercompose_project/`](resources/dockercompose_project/) — wrap an existing `docker-compose.yml` or pass inline YAML
- [`data-sources/dockercompose_project/`](data-sources/dockercompose_project/) — read runtime state of an existing project

## Running an example

```bash
cd examples/wordpress
terraform init
terraform apply
# ... use it ...
terraform destroy
```

> All examples use Docker Compose locally by default. For remote hosts, set `DOCKER_HOST=ssh://user@server` or configure the `host` attribute in the `provider` block — see [`remote-vps/`](remote-vps/) for the full pattern.

## Verification policy

Every end-to-end example in this directory was actually applied against a live Docker daemon, the running services were probed with `curl`, and the stack was destroyed cleanly. The verification line in each README states the date, Terraform version, and Docker version it was last tested with.

The `remote-vps/` example is the only one not fully applied in CI — it requires an actual VPS — but `terraform validate` and `terraform plan` pass, and the underlying transport (`docker --host ssh://...`) is the standard Docker CLI mechanism.

## More

- [Provider docs on Terraform Registry](https://registry.terraform.io/providers/xRizur/dockercompose/latest/docs)
- [Top-level README with full reference](../README.md)
- [Issues / questions](https://github.com/xRizur/terraform-provider-dockercompose/issues)
