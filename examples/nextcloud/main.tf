terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.1"
    }
  }
}

provider "dockercompose" {}

variable "nextcloud_port" {
  description = "Host port for Nextcloud."
  type        = number
  default     = 8081
}

variable "db_password" {
  description = "Postgres password for the nextcloud user."
  type        = string
  default     = "change-me-pg"
  sensitive   = true
}

variable "admin_user" {
  description = "Initial Nextcloud admin username."
  type        = string
  default     = "admin"
}

variable "admin_password" {
  description = "Initial Nextcloud admin password."
  type        = string
  default     = "change-me-admin"
  sensitive   = true
}

variable "trusted_domains" {
  description = "Space-separated trusted domains for Nextcloud."
  type        = string
  default     = "localhost"
}

resource "dockercompose_stack" "nextcloud" {
  name                      = "nextcloud-demo"
  remove_volumes_on_destroy = true

  service {
    name           = "db"
    image          = "postgres:16-alpine"
    container_name = "nc_db"
    restart        = "unless-stopped"

    environment = {
      POSTGRES_DB       = "nextcloud"
      POSTGRES_USER     = "nextcloud"
      POSTGRES_PASSWORD = var.db_password
    }

    volumes  = ["nc_db_data:/var/lib/postgresql/data"]
    networks = ["nc_net"]

    healthcheck_test         = ["CMD-SHELL", "pg_isready -U nextcloud -d nextcloud"]
    healthcheck_interval     = "10s"
    healthcheck_timeout      = "5s"
    healthcheck_retries      = 10
    healthcheck_start_period = "20s"

    resource_limits_memory = "512M"
  }

  service {
    name           = "redis"
    image          = "redis:7-alpine"
    container_name = "nc_redis"
    restart        = "unless-stopped"

    networks = ["nc_net"]

    healthcheck_test     = ["CMD", "redis-cli", "ping"]
    healthcheck_interval = "10s"
    healthcheck_timeout  = "5s"
    healthcheck_retries  = 5

    resource_limits_memory = "128M"
  }

  service {
    name           = "app"
    image          = "nextcloud:30-apache"
    container_name = "nc_app"
    restart        = "unless-stopped"
    depends_on     = ["db", "redis"]

    ports = ["${var.nextcloud_port}:80"]

    environment = {
      POSTGRES_HOST          = "db"
      POSTGRES_DB            = "nextcloud"
      POSTGRES_USER          = "nextcloud"
      POSTGRES_PASSWORD      = var.db_password
      REDIS_HOST             = "redis"
      NEXTCLOUD_ADMIN_USER   = var.admin_user
      NEXTCLOUD_ADMIN_PASSWORD = var.admin_password
      NEXTCLOUD_TRUSTED_DOMAINS = var.trusted_domains
    }

    volumes = [
      "nc_app_data:/var/www/html",
      "nc_user_data:/var/www/html/data",
    ]

    networks = ["nc_net"]

    resource_limits_memory = "1G"
  }

  volume {
    name = "nc_db_data"
  }

  volume {
    name = "nc_app_data"
  }

  volume {
    name = "nc_user_data"
  }

  network {
    name   = "nc_net"
    driver = "bridge"
  }
}

output "nextcloud_url" {
  description = "Nextcloud URL. First load takes ~1-2 min while installer runs."
  value       = "http://localhost:${var.nextcloud_port}"
}

output "admin_credentials" {
  value = {
    username = var.admin_user
    password = var.admin_password
  }
  sensitive = true
}

output "container_health" {
  value = { for c in dockercompose_stack.nextcloud.container : c.service => c.health }
}
