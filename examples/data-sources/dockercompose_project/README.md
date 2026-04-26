# Example: `dockercompose_project` data source — read-only project lookup

Reads the runtime state of an existing Docker Compose project (one started by `dockercompose_stack`, `dockercompose_project`, or even `docker compose up` outside Terraform). Useful for:

- Wiring container IPs / ports into other Terraform resources (DNS records, load balancer targets, monitoring config)
- Asserting project health from `terraform plan` / external checks
- Cross-stack composition: one Terraform config creates the stack, another reads its outputs

## Outputs in this example

| Output | What it shows |
|---|---|
| `project_status` | Overall project state (`running`, `partial`, `stopped`) |
| `web_container_ip` | IP address of the first container in the project |
| `container_count` | Total number of containers |

## Run

This example assumes a project named `backend-services` already exists. If you don't have one, run the [`dockercompose_project` resource example](../../resources/dockercompose_project/) first.

```bash
terraform init
terraform apply
terraform output
```

Because it's read-only, `terraform destroy` is a no-op for the data source — it doesn't tear anything down.

## Tip: per-service lookups

To grab a specific service's IP rather than `[0]`:

```hcl
output "db_ip" {
  value = [for c in data.dockercompose_project.backend.container : c.ip_address if c.service == "database"][0]
}
```

## Related

- [Full attribute reference](https://registry.terraform.io/providers/xRizur/dockercompose/latest/docs/data-sources/project)
