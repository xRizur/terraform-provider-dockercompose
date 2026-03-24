terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.1"
    }
  }
}

# Look up an existing Docker Compose project by name
data "dockercompose_project" "backend" {
  name = "backend-services"
}

# Use the data source outputs
output "project_status" {
  description = "Overall project status"
  value       = data.dockercompose_project.backend.status
}

output "web_container_ip" {
  description = "IP address of the first container"
  value       = data.dockercompose_project.backend.container[0].ip_address
}

output "container_count" {
  description = "Total number of containers in the project"
  value       = length(data.dockercompose_project.backend.container)
}
