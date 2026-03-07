terraform {
  required_providers {
    dockercompose = {
      source = "macie/dockercompose"
    }
  }
}

# ──────────────────────────────────────────────
# Provider configuration
# ──────────────────────────────────────────────

# Local Docker daemon (default)
provider "dockercompose" {}

# Remote host examples:
# provider "dockercompose" {
#   host = "ssh://deploy@production-server"
# }
# provider "dockercompose" {
#   host          = "tcp://docker.example.com:2376"
#   docker_binary = "/usr/local/bin/docker"
# }

# ──────────────────────────────────────────────
# Example 1: Full HCL-defined stack
# All images are public and pullable without auth.
# ──────────────────────────────────────────────

resource "dockercompose_stack" "myapp" {
  name                      = "myapp"
  remove_volumes_on_destroy = true

  # --- Web (reverse proxy) ---
  service {
    name       = "nginx"
    image      = "nginx:alpine"
    restart    = "unless-stopped"
    ports      = ["8080:80"]
    depends_on = ["api"]
    networks   = ["frontend", "backend"]

    healthcheck_test     = ["CMD", "curl", "-sf", "http://localhost"]
    healthcheck_interval = "30s"
    healthcheck_retries  = 3

    labels = {
      "com.example.description" = "Web frontend"
    }
  }

  # --- API (Node.js) ---
  service {
    name       = "api"
    image      = "node:22-alpine"
    restart    = "unless-stopped"
    expose     = ["3000"]
    depends_on = ["db", "redis"]
    networks   = ["frontend", "backend"]
    command    = ["sh", "-c", "echo 'API  placeholder running' && sleep infinity"]

    environment = {
      DATABASE_URL = "postgres://admin:secret@db:5432/myapp"
      REDIS_URL    = "redis://redis:6379"
      NODE_ENV     = "production"
    }

    resource_limits_cpus   = "0.5"
    resource_limits_memory = "512M"

    logging_driver = "json-file"
    logging_options = {
      "max-size" = "10m"
      "max-file" = "3"
    }
  }

  # --- Database ---
  service {
    name     = "db"
    image    = "postgres:17-alpine"
    restart  = "always"
    ports    = ["5432:5432"]
    networks = ["backend"]

    environment = {
      POSTGRES_USER     = "admin"
      POSTGRES_PASSWORD = "secret"
      POSTGRES_DB       = "myapp"
    }

    volumes = ["pgdata:/var/lib/postgresql/data"]

    healthcheck_test     = ["CMD-SHELL", "pg_isready -U admin"]
    healthcheck_interval = "10s"
    healthcheck_retries  = 5

    shm_size = "256mb"
  }

  # --- Redis ---
  service {
    name     = "redis"
    image    = "redis:7-alpine"
    restart  = "always"
    networks = ["backend"]

    volumes = ["redis_data:/data"]

    healthcheck_test     = ["CMD", "redis-cli", "ping"]
    healthcheck_interval = "10s"
  }

  # --- Networks ---
  network {
    name   = "frontend"
    driver = "bridge"
  }

  network {
    name     = "backend"
    driver   = "bridge"
    internal = true

    ipam_driver  = "default"
    ipam_subnet  = "172.28.0.0/16"
    ipam_gateway = "172.28.0.1"
  }

  # --- Volumes ---
  volume {
    name = "pgdata"
  }

  volume {
    name = "redis_data"
  }
}

# ──────────────────────────────────────────────
# Example 2: Minimal single-service stack
# ──────────────────────────────────────────────

# resource "dockercompose_stack" "simple" {
#   name = "simple"
#   service {
#     name  = "web"
#     image = "nginx:alpine"
#     ports = ["9090:80"]
#   }
# }

# ──────────────────────────────────────────────
# Example 3: Raw YAML file reference
# ──────────────────────────────────────────────

# resource "dockercompose_project" "legacy" {
#   name         = "legacy-app"
#   compose_file = "${path.module}/docker-compose.yml"
# }

# ──────────────────────────────────────────────
# Example 4: Inline YAML with template variables
# ──────────────────────────────────────────────

# variable "image_tag" {
#   default = "alpine"
# }
#
# resource "dockercompose_project" "templated" {
#   name = "templated-app"
#   compose_yaml = <<-EOT
#     services:
#       app:
#         image: nginx:${var.image_tag}
#         ports:
#           - "9091:80"
#   EOT
# }

# ──────────────────────────────────────────────
# Outputs
# ──────────────────────────────────────────────

output "compose_yaml" {
  value     = dockercompose_stack.myapp.compose_yaml
  sensitive = true
}

output "compose_file_path" {
  value = dockercompose_stack.myapp.compose_file_path
}

# ── Container runtime attributes ─────────────
# After apply, each container's runtime info is available.

output "all_containers" {
  description = "Full container info list (service, ID, name, state, IP, ports, networks)"
  value       = dockercompose_stack.myapp.container
}

output "nginx_container_id" {
  description = "Container ID of the nginx service"
  value       = [for c in dockercompose_stack.myapp.container : c.container_id if c.service == "nginx"][0]
}

output "nginx_ip" {
  description = "IP address of the nginx container (first network)"
  value       = [for c in dockercompose_stack.myapp.container : c.ip_address if c.service == "nginx"][0]
}

output "db_health" {
  description = "Health status of the db service"
  value       = [for c in dockercompose_stack.myapp.container : c.health if c.service == "db"][0]
}

output "container_ips" {
  description = "Map of service name → IP address"
  value       = { for c in dockercompose_stack.myapp.container : c.service => c.ip_address }
}

output "nginx_network_details" {
  description = "All network settings for the nginx service"
  value       = [for c in dockercompose_stack.myapp.container : c.network_settings if c.service == "nginx"][0]
}

output "published_ports" {
  description = "All containers with published (public) ports"
  value = {
    for c in dockercompose_stack.myapp.container : c.service => [
      for p in c.ports : "${p.ip}:${p.public_port}->${p.private_port}/${p.protocol}"
      if p.public_port > 0
    ] if length([for p in c.ports : p if p.public_port > 0]) > 0
  }
}