# Terraform Provider for Docker Compose

A Terraform provider that manages Docker Compose stacks - define multi-container applications in HCL with full lifecycle management, remote host support, and comprehensive Docker Compose spec coverage.

## Features

- **Remote host support** - connect via SSH, TCP, or Unix socket (like the Docker provider)
- **Two resource types**:
  - `dockercompose_stack` - full HCL-modeled services, networks, volumes, configs, secrets
  - `dockercompose_project` - use existing `docker-compose.yml` files or inline YAML
- **Comprehensive service config** - ports, volumes, environment, healthchecks, deploy resources, logging, security options, sysctls, devices, and 50+ other Docker Compose fields
- **Network & volume management** - drivers, IPAM, external references, labels, driver options
- **Docker configs & secrets** - top-level config/secret definitions
- **Project isolation** - each stack gets its own project name (`-p`) and directory
- **State management** - generated YAML stored in Terraform state, auto-restored if deleted from disk
- **Import support** - import existing stacks with `terraform import`

## Quick Start

```hcl
terraform {
  required_providers {
    dockercompose = {
      source = "xRizur/dockercompose"
    }
  }
}

provider "dockercompose" {}

resource "dockercompose_stack" "app" {
  name = "myapp"

  service {
    name    = "web"
    image   = "nginx:alpine"
    restart = "unless-stopped"
    ports   = ["8080:80"]
  }

  service {
    name    = "db"
    image   = "postgres:17-alpine"
    restart = "always"
    ports   = ["5432:5432"]
    volumes = ["pgdata:/var/lib/postgresql/data"]
    environment = {
      POSTGRES_PASSWORD = "secret"
    }
    healthcheck_test     = ["CMD-SHELL", "pg_isready"]
    healthcheck_interval = "10s"
  }

  volume {
    name = "pgdata"
  }
}
```

## Provider Configuration

```hcl
provider "dockercompose" {
  # Connect to remote Docker host (optional)
  host = "ssh://deploy@production-server"
  # host = "tcp://docker.example.com:2376"
  # host = "unix:///var/run/docker.sock"

  # Custom docker binary path (default: "docker")
  docker_binary = "/usr/local/bin/docker"

  # Directory for generated compose files (default: ~/.terraform-docker-compose)
  project_directory = "/opt/compose-projects"
}
```

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `host` | string | `$DOCKER_HOST` | Docker daemon URL (ssh://, tcp://, unix://) |
| `docker_binary` | string | `"docker"` | Path to docker binary |
| `project_directory` | string | `~/.terraform-docker-compose` | Base directory for compose files |

## Resources

### `dockercompose_stack`

Full HCL-modeled Docker Compose stack with typed attributes.

#### Top-level Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Project name (ForceNew) |
| `working_dir` | string | no | Working directory for relative paths |
| `remove_volumes_on_destroy` | bool | no | Run `docker compose down -v` on destroy |

#### Computed Outputs

| Attribute | Description |
|-----------|-------------|
| `compose_yaml` | Generated YAML content |
| `compose_file_path` | Path to generated file |
| `container` | List of container runtime info (see below) |

#### Container Block (computed)

After `apply`, every running container is exposed as a computed `container` list sorted by service name.
Use `for` expressions to look up specific services.

| Attribute | Type | Description |
|-----------|------|-------------|
| `service` | string | Service name |
| `container_id` | string | Short container ID (12 chars) |
| `container_name` | string | Container name (e.g. `myapp-web-1`) |
| `image` | string | Docker image |
| `state` | string | `running`, `exited`, `paused`, etc. |
| `health` | string | `healthy`, `unhealthy`, `starting`, or empty |
| `exit_code` | int | Exit code (0 when running) |
| `ip_address` | string | IP on the first attached network |
| `ports` | list | Published port mappings (see sub-table) |
| `network_settings` | list | Per-network details (see sub-table) |

**`ports` sub-attributes:**

| Attribute | Type | Description |
|-----------|------|-------------|
| `ip` | string | Bound host IP (e.g. `0.0.0.0`) |
| `private_port` | int | Container-side port |
| `public_port` | int | Host-side port (0 if expose-only) |
| `protocol` | string | `tcp` or `udp` |

**`network_settings` sub-attributes:**

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Docker network name |
| `ip_address` | string | Container IP on this network |
| `gateway` | string | Network gateway |
| `mac_address` | string | Container MAC address |

**Example outputs:**

```hcl
# Single service lookup
output "web_ip" {
  value = [for c in dockercompose_stack.app.container : c.ip_address if c.service == "web"][0]
}

# All IPs as a map
output "container_ips" {
  value = { for c in dockercompose_stack.app.container : c.service => c.ip_address }
}

# Published ports
output "published_ports" {
  value = {
    for c in dockercompose_stack.app.container : c.service => [
      for p in c.ports : "${p.ip}:${p.public_port}->${p.private_port}/${p.protocol}"
      if p.public_port > 0
    ] if length([for p in c.ports : p if p.public_port > 0]) > 0
  }
}

# Network details for a specific service
output "db_networks" {
  value = [for c in dockercompose_stack.app.container : c.network_settings if c.service == "db"][0]
}

# Container ID (useful with docker_exec or provisioners)
output "web_container_id" {
  value = [for c in dockercompose_stack.app.container : c.container_id if c.service == "web"][0]
}
```

#### Service Block

```hcl
service {
  # Core
  name           = "api"
  image          = "myapp:latest"
  container_name = "myapp-api"
  restart        = "unless-stopped"    # no, always, on-failure, unless-stopped
  ports          = ["3000:3000"]
  expose         = ["9090"]
  depends_on     = ["db", "redis"]
  environment    = { NODE_ENV = "production" }
  env_file       = [".env", ".env.production"]
  command        = ["node", "server.js"]
  entrypoint     = ["/entrypoint.sh"]
  volumes        = ["./data:/data", "dbvol:/db"]
  networks       = ["frontend", "backend"]
  labels         = { "com.example.team" = "backend" }

  # Deploy
  replicas                     = 3
  resource_limits_cpus         = "0.5"
  resource_limits_memory       = "512M"
  resource_reservations_cpus   = "0.25"
  resource_reservations_memory = "256M"

  # Healthcheck
  healthcheck_test         = ["CMD", "curl", "-f", "http://localhost:3000/health"]
  healthcheck_interval     = "30s"
  healthcheck_timeout      = "10s"
  healthcheck_retries      = 3
  healthcheck_start_period = "40s"
  healthcheck_disable      = false

  # Logging
  logging_driver  = "json-file"
  logging_options  = { "max-size" = "10m", "max-file" = "3" }

  # Security
  cap_add      = ["NET_ADMIN"]
  cap_drop     = ["ALL"]
  security_opt = ["no-new-privileges:true"]
  privileged   = false
  read_only    = true
  init         = true
  user         = "1000:1000"

  # Networking
  dns          = ["8.8.8.8"]
  extra_hosts  = ["host.docker.internal:host-gateway"]
  hostname     = "api"
  domainname   = "example.com"
  network_mode = ""         # bridge, host, none, service:name

  # Runtime
  working_dir       = "/app"
  stdin_open        = false
  tty               = false
  shm_size          = "256m"
  stop_grace_period = "30s"
  stop_signal       = "SIGTERM"
  platform          = "linux/amd64"
  pull_policy       = "always"    # always, never, missing, build
  runtime           = "runc"
  tmpfs             = ["/tmp"]
  devices           = ["/dev/sda:/dev/xvdc:rwm"]
  sysctls           = { "net.core.somaxconn" = "1024" }
  profiles          = ["debug"]
  pid               = ""
  ipc               = ""

  # Legacy resource limits
  mem_limit       = ""
  mem_reservation = ""
  cpus            = ""
}
```

#### Network Block

```hcl
network {
  name        = "backend"
  driver      = "bridge"
  driver_opts = { "com.docker.network.bridge.enable_icc" = "true" }
  external    = false
  internal    = true
  attachable  = false
  labels      = { "env" = "production" }

  # IPAM
  ipam_driver  = "default"
  ipam_subnet  = "172.28.0.0/16"
  ipam_gateway = "172.28.0.1"
}
```

#### Volume Block

```hcl
volume {
  name        = "dbdata"
  driver      = "local"
  driver_opts = { "type" = "nfs", "o" = "addr=192.168.1.1", "device" = ":/data" }
  external    = false
  labels      = { "backup" = "daily" }
}
```

#### Config & Secret Blocks

```hcl
config {
  name = "nginx_config"
  file = "./nginx.conf"
}

secret {
  name = "db_password"
  file = "./secrets/db_pass.txt"
}
```

### `dockercompose_project`

Manages a stack from an existing `docker-compose.yml` file or inline YAML.

```hcl
# From file
resource "dockercompose_project" "legacy" {
  name         = "legacy-app"
  compose_file = "${path.module}/docker-compose.yml"
}

# Inline YAML (supports templatefile)
resource "dockercompose_project" "dynamic" {
  name = "dynamic-app"
  compose_yaml = templatefile("${path.module}/compose.yml.tpl", {
    image_tag   = var.image_tag
    db_password = var.db_password
  })
}
```

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Project name (required, ForceNew) |
| `compose_file` | string | Path to compose file (conflicts with compose_yaml) |
| `compose_yaml` | string | Inline YAML content (conflicts with compose_file) |
| `remove_volumes_on_destroy` | bool | Remove volumes on destroy |
| `yaml_sha256` | string | (computed) SHA256 of YAML content |
| `container` | list | (computed) Container runtime info - same schema as `dockercompose_stack` |

## How It Works

1. **Create/Update**: Builds YAML from HCL config (or uses provided YAML), writes to `<project_directory>/<name>/docker-compose.yml`, runs `docker compose -p <name> up -d --remove-orphans`
2. **Read**: Checks if compose file exists (restores from state if missing), verifies services are running via `docker compose ps --services`
3. **Delete**: Runs `docker compose -p <name> down [-v]`, cleans up generated files
4. **State**: Generated YAML stored in Terraform state for recovery. Each stack isolated by project name and directory.

## Architecture

```
main.go              Entry point
provider.go          Provider schema + configuration (host, binary, directory)
docker.go            DockerClient - CLI wrapper with DOCKER_HOST support
compose.go           ComposeFile structs + YAML marshal/unmarshal
container_info.go    Container runtime schema, JSON parsing, readContainerInfo
resource_stack.go    dockercompose_stack - HCL-to-YAML resource
resource_project.go  dockercompose_project - raw YAML/file resource
utils.go             Type-safe extraction helpers
```

## Building

```bash
go build -o terraform-provider-dockercompose      # Linux / macOS
go build -o terraform-provider-dockercompose.exe   # Windows
```

## Local Testing (dev override)

The fastest way to test the provider locally - no registry, no `terraform init`.

### Prerequisites

- **Go** >= 1.21
- **Terraform** >= 1.0
- **Docker Desktop** (or Docker Engine with the Compose plugin)
- **Windows only**: Developer Mode enabled (Settings → System → For developers)

### 1. Build the provider binary

```bash
# From the repository root
go build -o terraform-provider-dockercompose.exe .   # Windows
go build -o terraform-provider-dockercompose .        # Linux / macOS
```

### 2. Create the Terraform CLI config with dev_overrides

**Windows** - create `%APPDATA%\terraform.rc` (typically `C:\Users\<you>\AppData\Roaming\terraform.rc`):

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/xRizur/dockercompose" = "C:\\Users\\<you>\\path\\to\\DockerCompose-Terraform-Provider"
  }
  direct {}
}
```

**Linux / macOS** - create `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/xRizur/dockercompose" = "/home/<you>/path/to/DockerCompose-Terraform-Provider"
  }
  direct {}
}
```

> The path must point to the directory containing the built binary.

### 3. Write a Terraform config

```hcl
terraform {
  required_providers {
    dockercompose = {
      source = "xRizur/dockercompose"
    }
  }
}

provider "dockercompose" {}

resource "dockercompose_stack" "demo" {
  name = "demo"
  service {
    name  = "web"
    image = "nginx:alpine"
    ports = ["9090:80"]
  }
}
```

### 4. Run Terraform

```bash
# No 'terraform init' needed with dev_overrides!
terraform plan
terraform apply -auto-approve

# Verify
docker compose -p demo ps

# Clean up
terraform destroy -auto-approve
```

> The warning "Provider development overrides are in effect" is expected - it means Terraform is using your local build.

### 5. Running the test suite

```bash
# Unit + integration tests (no Docker required)
go test -v -run "Test[^A]" -count=1 ./...

# Full suite including acceptance tests (Docker required)
TF_ACC=1 go test -v -count=1 -timeout 10m ./...        # Linux / macOS
$env:TF_ACC = "1"; go test -v -count=1 -timeout 10m ./...  # Windows PowerShell

# Coverage report (opens in browser)
go test -run "Test[^A]" -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Requirements

- Docker Engine with the Compose plugin (`docker compose`)
- Go 1.24+ (for building)
- For remote hosts: SSH access configured or TLS certificates

## Breaking Changes from v1

- **Provider config**: New `host`, `docker_binary`, `project_directory` fields
- **Service block**: Changed from `TypeSet` to `TypeList` - service order in HCL matters
- **Healthcheck**: `healthcheck_test` is now a list (was a string)
- **Replicas**: Moved from `replicas` direct field to deploy-aware config
- **YAML generation**: Uses struct marshaling instead of Go templates - output format may differ
- **Project isolation**: Each stack now uses `-p name` for Docker Compose project isolation
- **New resource**: `dockercompose_project` for raw YAML workflows
- Existing v1 state must be destroyed and recreated

