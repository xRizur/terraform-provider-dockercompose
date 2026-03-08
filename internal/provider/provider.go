package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/xRizur/terraform-provider-dockercompose/internal/docker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns the Terraform provider for Docker Compose.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOCKER_HOST", ""),
				Description: "Docker daemon host URL. Supports ssh://user@host, tcp://host:2376, unix:///path/to/socket. Defaults to DOCKER_HOST env var.",
			},
			"docker_binary": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "docker",
				Description: "Path to the docker binary.",
			},
			"project_directory": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					home, err := os.UserHomeDir()
					if err != nil {
						return nil, err
					}
					return filepath.Join(home, ".terraform-docker-compose"), nil
				},
				Description: "Base directory for storing generated docker-compose files. Defaults to ~/.terraform-docker-compose.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"dockercompose_stack":   resourceComposeStack(),
			"dockercompose_project": resourceComposeProject(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := &docker.DockerClient{
		Host:             d.Get("host").(string),
		Binary:           d.Get("docker_binary").(string),
		ProjectDirectory: d.Get("project_directory").(string),
	}

	if err := os.MkdirAll(client.ProjectDirectory, 0755); err != nil {
		return nil, fmt.Errorf("error creating project directory: %s", err)
	}

	return client, nil
}
