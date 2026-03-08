terraform {
  required_providers {
    dockercompose = {
      source  = "xRizur/dockercompose"
      version = "~> 1.0"
    }
  }
}

resource "dockercompose_project" "legacy_app" {
  name = "backend-services"

  # Option 1: Point to an existing file on the disk
  compose_file = "/var/app/backend-services/docker-compose.yml"

  # Make sure we clean named volumes when tearing down the project via terraform destroy
  remove_volumes_on_destroy = true
}

resource "dockercompose_project" "yaml_app" {
  name = "inline-services"

  # Option 2: Provide the YAML directly within Terraform, very useful with templatefile()
  compose_yaml = <<-EOT
    version: '3'
    services:
      web:
        image: nginx:latest
        ports:
          - "80:80"
  EOT

  remove_volumes_on_destroy = true
}
