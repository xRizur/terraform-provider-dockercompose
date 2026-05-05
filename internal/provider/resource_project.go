package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xRizur/terraform-provider-dockercompose/internal/docker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceComposeProject provides a resource that manages existing or inline docker-compose files.
// Unlike dockercompose_stack which builds YAML from HCL blocks, this accepts raw YAML or a file path.
func resourceComposeProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: importProject,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project name for the Docker Compose stack.",
			},
			"compose_file": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"compose_yaml"},
				Description:   "Path to an existing docker-compose.yml file.",
			},
			"compose_yaml": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"compose_file"},
				Description:   "Inline docker-compose YAML content. Supports templatefile().",
			},
			"remove_volumes_on_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to remove volumes on destroy.",
			},
			// Computed
			"yaml_sha256": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA256 hash of the compose YAML for change detection.",
			},

			// Per-resource host override (conflicts with connection)
			"host": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"ssh_connection"},
				Description:   "Docker daemon host URL. Overrides provider host when set. Conflicts with ssh_connection.",
			},

			// Per-resource SSH connection block (conflicts with host)
			"ssh_connection": connectionSchema(),

			// Container runtime info (populated after apply)
			"container": containerSchema(),
		},
	}
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	providerClient := m.(*docker.DockerClient)
	stackName := d.Get("name").(string)

	resourceHost := d.Get("host").(string)
	conn := connectionFromResourceData(d)
	host, err := docker.EffectiveHost(resourceHost, conn, providerClient.Host)
	if err != nil {
		return err
	}
	client := docker.ClientForHost(host, providerClient)

	composeFilePath, err := resolveProjectComposeFile(d, client, stackName)
	if err != nil {
		return err
	}

	if _, err := client.ComposeUp(stackName, composeFilePath); err != nil {
		return fmt.Errorf("error starting project: %s", err)
	}

	if resourceHost != "" || conn != nil {
		d.SetId(host + "/" + stackName)
	} else {
		d.SetId(stackName)
	}

	// Compute hash for change detection
	content, err := os.ReadFile(composeFilePath)
	if err == nil {
		hash := fmt.Sprintf("%x", sha256.Sum256(content))
		if setErr := d.Set("yaml_sha256", hash); setErr != nil {
			return fmt.Errorf("error setting yaml_sha256: %s", setErr)
		}
	}

	return resourceProjectRead(d, m)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	providerClient := m.(*docker.DockerClient)

	resourceHost := d.Get("host").(string)
	conn := connectionFromResourceData(d)
	host, err := docker.EffectiveHost(resourceHost, conn, providerClient.Host)
	if err != nil {
		d.SetId("")
		return nil
	}
	client := docker.ClientForHost(host, providerClient)

	// Derive the stack name from the resource ID (strip optional host prefix)
	id := d.Id()
	stackName := id
	if idx := strings.LastIndex(id, "/"); idx > 0 {
		stackName = id[idx+1:]
	}

	composeFilePath, err := resolveProjectComposeFile(d, client, stackName)
	if err != nil {
		d.SetId("")
		return nil
	}

	output, err := client.ComposePSServices(stackName, composeFilePath)
	if err != nil || strings.TrimSpace(output) == "" {
		d.SetId("")
		return nil
	}

	// Read container runtime info (IDs, IPs, ports, health, etc.)
	return readContainerInfo(d, client, stackName, composeFilePath)
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceProjectCreate(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	providerClient := m.(*docker.DockerClient)

	resourceHost := d.Get("host").(string)
	conn := connectionFromResourceData(d)
	host, err := docker.EffectiveHost(resourceHost, conn, providerClient.Host)
	if err != nil {
		return err
	}
	client := docker.ClientForHost(host, providerClient)

	// Derive the stack name from the resource ID (strip optional host prefix)
	id := d.Id()
	stackName := id
	if idx := strings.LastIndex(id, "/"); idx > 0 {
		stackName = id[idx+1:]
	}

	removeVolumes := d.Get("remove_volumes_on_destroy").(bool)

	composeFilePath, err := resolveProjectComposeFile(d, client, stackName)
	if err != nil {
		return err
	}

	if _, err := client.ComposeDown(stackName, composeFilePath, removeVolumes); err != nil {
		return fmt.Errorf("error stopping project: %s", err)
	}

	// Clean up generated file (only if we wrote one from inline YAML)
	if _, ok := d.GetOk("compose_yaml"); ok {
		os.Remove(composeFilePath)
		os.Remove(filepath.Dir(composeFilePath))
	}

	return nil
}

// importProject handles terraform import for dockercompose_project.
// Accepts IDs in two forms:
//   - "projectName"                   → uses provider host
//   - "ssh://user@host:port/projName" → sets resource-level host
func importProject(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	id := d.Id()
	if idx := strings.LastIndex(id, "/"); idx > 0 {
		host := id[:idx]
		name := id[idx+1:]
		if err := d.Set("host", host); err != nil {
			return nil, fmt.Errorf("error setting host during import: %s", err)
		}
		if err := d.Set("name", name); err != nil {
			return nil, fmt.Errorf("error setting name during import: %s", err)
		}
		d.SetId(id)
	} else {
		if err := d.Set("name", id); err != nil {
			return nil, fmt.Errorf("error setting name during import: %s", err)
		}
		d.SetId(id)
	}
	return []*schema.ResourceData{d}, nil
}


// resolveProjectComposeFile determines the compose file path from either compose_file or compose_yaml.
func resolveProjectComposeFile(d *schema.ResourceData, client *docker.DockerClient, stackName string) (string, error) {
	// Option 1: explicit file path
	if f, ok := d.GetOk("compose_file"); ok && f.(string) != "" {
		path := f.(string)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", fmt.Errorf("compose file not found: %s", path)
		}
		return path, nil
	}

	// Option 2: inline YAML content
	if y, ok := d.GetOk("compose_yaml"); ok && y.(string) != "" {
		projectDir := client.ProjectDir(stackName)
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return "", fmt.Errorf("error creating project directory: %s", err)
		}
		composeFilePath := filepath.Join(projectDir, "docker-compose.yml")
		if err := os.WriteFile(composeFilePath, []byte(y.(string)), 0644); err != nil {
			return "", fmt.Errorf("error writing compose file: %s", err)
		}
		return composeFilePath, nil
	}

	return "", fmt.Errorf("either 'compose_file' or 'compose_yaml' must be specified")
}
