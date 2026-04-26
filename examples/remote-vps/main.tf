terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.1"
    }
  }
}

# ============================================================
# Remote VPS deployment via SSH
# ============================================================
#
# This example deploys a Docker Compose stack to a *remote* host
# without installing Terraform or this provider on the VPS.
# Everything runs on your laptop; the provider invokes
# `docker --host ssh://...` against the remote Docker daemon.
#
# REQUIREMENTS on the local machine:
#   - Docker CLI 19.03+ (for ssh:// transport support)
#   - SSH client + an SSH key already authorized on the VPS
#     (the provider does not handle interactive password prompts)
#
# REQUIREMENTS on the remote VPS:
#   - Docker Engine + Compose plugin installed
#   - Your SSH user must be in the `docker` group (or be root)
#   - Outbound network access to pull images
#
# QUICK SANITY CHECK before running terraform:
#   docker --host ssh://deploy@vps.example.com info
# If that succeeds, this example will work too.
# ============================================================

variable "ssh_host" {
  description = "VPS connection string (user@host or user@host:port)."
  type        = string
  # Example: "deploy@vps.example.com"
}

variable "app_port" {
  description = "Port to expose on the VPS (firewall must allow it)."
  type        = number
  default     = 80
}

variable "remote_project_dir" {
  description = "Where to store generated compose files on the VPS."
  type        = string
  default     = "/opt/terraform-compose"
}

provider "dockercompose" {
  host              = "ssh://${var.ssh_host}"
  project_directory = var.remote_project_dir
}

resource "dockercompose_stack" "edge_app" {
  name                      = "edge-app"
  remove_volumes_on_destroy = false

  service {
    name    = "web"
    image   = "nginx:alpine"
    restart = "unless-stopped"

    ports = ["${var.app_port}:80"]

    volumes  = ["web_content:/usr/share/nginx/html"]
    networks = ["edge_net"]

    healthcheck_test     = ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/"]
    healthcheck_interval = "30s"
    healthcheck_timeout  = "5s"
    healthcheck_retries  = 3

    resource_limits_memory = "128M"
  }

  volume {
    name = "web_content"
  }

  network {
    name   = "edge_net"
    driver = "bridge"
  }
}

output "app_url" {
  description = "Public URL (assumes the VPS hostname resolves and the firewall allows app_port)."
  value       = "http://${split("@", var.ssh_host)[1]}:${var.app_port}"
}

output "container_status" {
  description = "Runtime info pulled from the remote daemon."
  value = {
    for c in dockercompose_stack.edge_app.container :
    c.service => {
      state  = c.state
      health = c.health
      image  = c.image
    }
  }
}
