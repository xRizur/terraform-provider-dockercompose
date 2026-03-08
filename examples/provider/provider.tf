terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.0"
    }
  }
}

# Default configuration using local Docker instance
provider "dockercompose" {
  # (Optional) Defaults to the DOCKER_HOST environment variable if omitted.
  # host = "unix:///var/run/docker.sock"

  # (Optional) Provide a specific path to the docker executable.
  # docker_binary = "/usr/local/bin/docker"
  
  # (Optional) Override the base directory where docker-compose files and stack context will be saved.
  # project_directory = "/var/lib/terraform-docker-compose"
}
