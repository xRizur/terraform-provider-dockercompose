terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.0"
    }
  }
}

resource "dockercompose_stack" "monitoring" {
  name = "metrics-stack"

  # Clean up associated volumes strictly on destroy
  remove_volumes_on_destroy = true

  # Primary database service
  service {
    name           = "database"
    image          = "postgres:15-alpine"
    container_name = "monitoring_db"
    restart        = "unless-stopped"

    environment = {
      POSTGRES_USER     = "admin"
      POSTGRES_PASSWORD = "secretpassword"
      POSTGRES_DB       = "metrics"
    }

    # Restrict Hardware resources
    resource_limits_memory = "512M"
    resource_limits_cpus   = "1.0"

    # Volume mounts are defined as strings (host:container or named:container)
    volumes = ["db_data:/var/lib/postgresql/data"]
    
    ports = ["5432:5432"]

    # Healthcheck attributes
    healthcheck_test     = ["CMD", "pg_isready", "-U", "admin"]
    healthcheck_interval = "10s"
    healthcheck_timeout  = "5s"
    healthcheck_retries  = 5
    
    networks = ["monitoring_net"]
  }

  # Web application
  service {
    name           = "grafana"
    image          = "grafana/grafana:latest"
    container_name = "monitoring_ui"
    depends_on     = ["database"]
    ports          = ["3000:3000"]
    
    volumes  = ["grafana_data:/var/lib/grafana"]
    networks = ["monitoring_net"]
  }

  # Register volume definitions
  volume {
    name = "db_data"
  }

  volume {
    name = "grafana_data"
  }

  # Register network definitions
  network {
    name = "monitoring_net"
  }
}
