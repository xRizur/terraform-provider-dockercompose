terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.1"
    }
  }
}

provider "dockercompose" {}

variable "wordpress_port" {
  description = "Host port to expose WordPress on."
  type        = number
  default     = 8080
}

variable "db_root_password" {
  description = "MySQL root password."
  type        = string
  default     = "change-me-root"
  sensitive   = true
}

variable "db_password" {
  description = "MySQL password for the wordpress user."
  type        = string
  default     = "change-me-wp"
  sensitive   = true
}

resource "dockercompose_stack" "wordpress" {
  name                      = "wordpress-demo"
  remove_volumes_on_destroy = true

  service {
    name           = "db"
    image          = "mysql:8.0"
    container_name = "wp_db"
    restart        = "unless-stopped"

    environment = {
      MYSQL_ROOT_PASSWORD = var.db_root_password
      MYSQL_DATABASE      = "wordpress"
      MYSQL_USER          = "wordpress"
      MYSQL_PASSWORD      = var.db_password
    }

    volumes  = ["wp_db_data:/var/lib/mysql"]
    networks = ["wp_net"]

    healthcheck_test         = ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uroot", "-p${var.db_root_password}"]
    healthcheck_interval     = "10s"
    healthcheck_timeout      = "5s"
    healthcheck_retries      = 10
    healthcheck_start_period = "30s"

    resource_limits_memory = "512M"
  }

  service {
    name           = "wordpress"
    image          = "wordpress:6-php8.2-apache"
    container_name = "wp_app"
    restart        = "unless-stopped"
    depends_on     = ["db"]

    ports = ["${var.wordpress_port}:80"]

    environment = {
      WORDPRESS_DB_HOST     = "db:3306"
      WORDPRESS_DB_NAME     = "wordpress"
      WORDPRESS_DB_USER     = "wordpress"
      WORDPRESS_DB_PASSWORD = var.db_password
    }

    volumes  = ["wp_files:/var/www/html"]
    networks = ["wp_net"]

    resource_limits_memory = "512M"
  }

  volume {
    name = "wp_db_data"
  }

  volume {
    name = "wp_files"
  }

  network {
    name   = "wp_net"
    driver = "bridge"
  }
}

output "wordpress_url" {
  description = "Open this URL in a browser to finish WordPress setup."
  value       = "http://localhost:${var.wordpress_port}"
}

output "container_status" {
  description = "Per-service runtime info."
  value = {
    for c in dockercompose_stack.wordpress.container :
    c.service => {
      state  = c.state
      health = c.health
      ports  = [for p in c.ports : "${p.public_port}->${p.private_port}/${p.protocol}" if p.public_port > 0]
    }
  }
}
