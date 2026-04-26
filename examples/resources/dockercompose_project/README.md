# Example: `dockercompose_project` — existing or inline YAML

Use `dockercompose_project` when you already have a `docker-compose.yml` file (or want to template one) and don't want to rewrite it as HCL. Two modes shown here:

1. **`compose_file`** — point at an existing file on disk. Terraform manages the lifecycle (`up`, `down`), the YAML stays unchanged.
2. **`compose_yaml`** — pass YAML inline as a string. Combine with `templatefile()` to inject variables, secrets, or per-environment values.

`compose_file` and `compose_yaml` are mutually exclusive — pick one per resource.

## Run

```bash
terraform init
terraform apply
```

Both projects (`backend-services` from a file, `inline-services` from inline YAML) come up in parallel.

```bash
docker compose -p backend-services ps
docker compose -p inline-services ps
```

Tear them down:

```bash
terraform destroy
```

## When to use this vs `dockercompose_stack`

| You have... | Use |
|---|---|
| An existing `docker-compose.yml` you want to keep verbatim | `dockercompose_project` with `compose_file` |
| A YAML template with variables (env, image tag, secrets) | `dockercompose_project` with `compose_yaml = templatefile(...)` |
| A new stack you're building fresh in IaC | `dockercompose_stack` (typed HCL, easier diffs) |

## Related

- [`dockercompose_stack` example](../dockercompose_stack/) — same workload modeled as HCL instead of YAML
- [Full attribute reference](https://registry.terraform.io/providers/xRizur/dockercompose/latest/docs/resources/project)
