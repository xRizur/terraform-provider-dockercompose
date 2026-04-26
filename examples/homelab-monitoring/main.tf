terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.1"
    }
  }
}

provider "dockercompose" {}

variable "grafana_port" {
  description = "Host port for Grafana UI."
  type        = number
  default     = 3000
}

variable "prometheus_port" {
  description = "Host port for Prometheus UI."
  type        = number
  default     = 9090
}

variable "loki_port" {
  description = "Host port for Loki API."
  type        = number
  default     = 3100
}

variable "grafana_admin_password" {
  description = "Initial Grafana admin password."
  type        = string
  default     = "admin"
  sensitive   = true
}

resource "dockercompose_stack" "monitoring" {
  name                      = "homelab-monitoring"
  working_dir               = abspath(path.module)
  remove_volumes_on_destroy = true

  service {
    name           = "prometheus"
    image          = "prom/prometheus:v2.54.1"
    container_name = "homelab_prometheus"
    restart        = "unless-stopped"

    ports = ["${var.prometheus_port}:9090"]

    volumes = [
      "${abspath(path.module)}/prometheus.yml:/etc/prometheus/prometheus.yml:ro",
      "prometheus_data:/prometheus",
    ]

    networks = ["monitoring"]

    command = [
      "--config.file=/etc/prometheus/prometheus.yml",
      "--storage.tsdb.path=/prometheus",
      "--web.console.libraries=/usr/share/prometheus/console_libraries",
      "--web.console.templates=/usr/share/prometheus/consoles",
    ]

    healthcheck_test     = ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9090/-/healthy"]
    healthcheck_interval = "10s"
    healthcheck_timeout  = "5s"
    healthcheck_retries  = 5
  }

  service {
    name           = "loki"
    image          = "grafana/loki:3.2.0"
    container_name = "homelab_loki"
    restart        = "unless-stopped"

    ports = ["${var.loki_port}:3100"]

    volumes = [
      "${abspath(path.module)}/loki-config.yml:/etc/loki/local-config.yaml:ro",
      "loki_data:/loki",
    ]

    networks = ["monitoring"]

    command = ["-config.file=/etc/loki/local-config.yaml"]

    healthcheck_test     = ["CMD-SHELL", "wget --quiet --tries=1 --spider http://localhost:3100/ready || exit 1"]
    healthcheck_interval = "10s"
    healthcheck_timeout  = "5s"
    healthcheck_retries  = 10
  }

  service {
    name           = "grafana"
    image          = "grafana/grafana:11.3.0"
    container_name = "homelab_grafana"
    restart        = "unless-stopped"
    depends_on     = ["prometheus", "loki"]

    ports = ["${var.grafana_port}:3000"]

    environment = {
      GF_SECURITY_ADMIN_PASSWORD = var.grafana_admin_password
      GF_USERS_ALLOW_SIGN_UP     = "false"
      GF_INSTALL_PLUGINS         = ""
    }

    volumes  = ["grafana_data:/var/lib/grafana"]
    networks = ["monitoring"]

    healthcheck_test     = ["CMD-SHELL", "wget --quiet --tries=1 --spider http://localhost:3000/api/health || exit 1"]
    healthcheck_interval = "10s"
    healthcheck_timeout  = "5s"
    healthcheck_retries  = 5
  }

  volume {
    name = "prometheus_data"
  }

  volume {
    name = "loki_data"
  }

  volume {
    name = "grafana_data"
  }

  network {
    name   = "monitoring"
    driver = "bridge"
  }
}

output "grafana_url" {
  description = "Open in browser. Login: admin / <grafana_admin_password>."
  value       = "http://localhost:${var.grafana_port}"
}

output "prometheus_url" {
  value = "http://localhost:${var.prometheus_port}"
}

output "loki_url" {
  value = "http://localhost:${var.loki_port}"
}

output "container_health" {
  value = {
    for c in dockercompose_stack.monitoring.container :
    c.service => c.health
  }
}
