# Deploy to a Remote VPS via SSH (no Terraform on the server)

Deploy a Docker Compose stack to a remote VPS without installing Terraform, this provider, or anything else on the server. The provider talks to the remote Docker daemon over SSH using Docker CLI's built-in `ssh://` transport — same mechanism `docker --host ssh://user@host` uses.

**Use cases:**

- Single-node production deployments (one VPS, no Kubernetes / Swarm)
- Edge devices: Raspberry Pi, NUC, MikroTik containers, OpenWrt boxes
- Disposable per-branch staging on cheap VPS instances
- Migrating an existing VPS workload into IaC without bootstrapping Terraform on it

**What you get:**

- Provider configured to drive a remote Docker daemon via SSH
- A minimal nginx stack as the deployable example (replace with your real workload)
- Output that constructs the public URL from the SSH host
- A `terraform.tfvars.example` template

## Verified

⚠️ **Validate-only test on `2026-04-26`.** This example was confirmed to pass `terraform validate` and `terraform plan` (with a placeholder `ssh_host`). Full `terraform apply` requires an actual VPS and is **not** covered by automated testing here — but the underlying mechanism (`docker --host ssh://...`) is widely used in production. Sanity-check on your VPS before running:

```bash
docker --host ssh://deploy@vps.example.com info
```

If that succeeds, this example will too.

## Prerequisites

**On your laptop:**

- Docker CLI 19.03+ (provides `ssh://` transport)
- SSH client + a key already authorized on the VPS (the provider does **not** prompt for passwords)

**On the VPS:**

- Docker Engine + Compose plugin
- The SSH user is in the `docker` group (or is root)
- Outbound network access to pull images
- Firewall opens whatever port you set (`app_port`, default 80)

## Run

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars and set ssh_host = "deploy@your-vps"

terraform init
terraform plan       # verify the plan against the *remote* daemon
terraform apply      # actually deploy
terraform destroy    # tear it down
```

## How it works

```
   your laptop                      VPS
+---------------+   ssh://      +-------------+
| terraform     |  ────────►   | docker daemon |
| docker CLI    |              | docker compose|
| (this provider)              | nginx         |
+---------------+              +-------------+
```

The provider runs `docker --host ssh://deploy@vps compose -p edge-app up -d`. The compose YAML is generated locally from your HCL, written to the VPS at `remote_project_dir`, and `docker compose` operates on it from there. State lives in your Terraform backend (local or remote), not on the VPS — the VPS only stores the compose file.

## Customize

```bash
# Custom SSH port
terraform apply -var='ssh_host=deploy@vps.example.com:2222'

# Different exposed port (don't forget the firewall)
terraform apply -var='app_port=8080'

# Different remote project directory
terraform apply -var='remote_project_dir=/srv/compose'
```

## Common gotchas

- **`docker info` over SSH hangs** — your SSH agent / key isn't loaded. Try `ssh-add` first, then re-run.
- **Permission denied on the docker socket** — your SSH user isn't in the `docker` group. `sudo usermod -aG docker $USER` on the VPS, then log out / back in.
- **Image pull is slow** — pulls happen on the VPS, not your laptop. Check VPS bandwidth, not yours.
- **Firewall blocks the published port** — `dockercompose_stack` only manages Docker; firewall rules on the VPS (UFW, firewalld, cloud security groups) are separate.

## Going further

- **Multiple environments**: one workspace per VPS, same config, different tfvars
- **Real workload**: replace the nginx service with your app (or use `dockercompose_project` to point at an existing `docker-compose.yml` on the VPS)
- **TLS**: add Caddy or Traefik as a service; Caddy auto-provisions Let's Encrypt certs with one line of config

## Keywords

terraform remote docker host, terraform docker over ssh, deploy docker compose to vps with terraform, terraform remote docker daemon, IaC single-node deployment, docker ssh transport terraform
