package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/xRizur/terraform-provider-dockercompose/internal/docker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceComposeProject provides a read-only data source that fetches
// runtime information about an existing Docker Compose project by name.
func dataSourceComposeProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectReadContext,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Docker Compose project name to look up.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Overall project status: running, partial, or stopped.",
			},
			"container": containerSchema(),
		},
	}
}

// dataSourceProjectReadContext queries the Docker engine for containers belonging to
// the given project and populates the data source attributes.
func dataSourceProjectReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*docker.DockerClient)
	projectName := d.Get("name").(string)

	// Query containers via docker compose ps --format json
	psJSON, err := client.ComposePSJSON(projectName, "")
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying project %q: %s", projectName, err))
	}

	entries, err := parseComposePSJSON(psJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing container info for project %q: %s", projectName, err))
	}

	if len(entries) == 0 {
		return diag.Errorf("no containers found for project %q", projectName)
	}

	// Set ID so Terraform tracks this data source instance
	d.SetId(projectName)

	// Determine and set overall project status
	status := determineProjectStatus(entries)
	if err := d.Set("status", status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %s", err))
	}

	// Collect container names for docker inspect (network details)
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}

	// Docker inspect to get network details (IPs, gateways, MACs)
	inspectMap := make(map[string]*DockerInspectEntry)
	if inspectJSON, err := client.DockerInspect(names...); err == nil {
		if inspected, err := parseDockerInspect(inspectJSON); err == nil {
			for i := range inspected {
				shortID := inspected[i].ID
				if len(shortID) > 12 {
					shortID = shortID[:12]
				}
				inspectMap[shortID] = &inspected[i]
			}
		}
	}

	// Sort entries by service name for deterministic ordering
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Service == entries[j].Service {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Service < entries[j].Service
	})

	// Build container list for Terraform state
	containers := make([]interface{}, len(entries))
	for i, e := range entries {
		c := map[string]interface{}{
			"service":        e.Service,
			"container_id":   e.ID,
			"container_name": e.Name,
			"image":          e.Image,
			"state":          e.State,
			"health":         e.Health,
			"exit_code":      e.ExitCode,
			"ip_address":     "",
		}

		// Port mappings
		ports := make([]interface{}, 0, len(e.Publishers))
		for _, p := range e.Publishers {
			ports = append(ports, map[string]interface{}{
				"ip":           p.URL,
				"private_port": p.TargetPort,
				"public_port":  p.PublishedPort,
				"protocol":     p.Protocol,
			})
		}
		c["ports"] = ports

		// Network settings from docker inspect
		networkSettings := []interface{}{}
		if insp, ok := inspectMap[e.ID]; ok {
			netNames := make([]string, 0, len(insp.NetworkSettings.Networks))
			for name := range insp.NetworkSettings.Networks {
				netNames = append(netNames, name)
			}
			sort.Strings(netNames)

			for idx, name := range netNames {
				net := insp.NetworkSettings.Networks[name]
				ns := map[string]interface{}{
					"name":        name,
					"ip_address":  net.IPAddress,
					"gateway":     net.Gateway,
					"mac_address": net.MacAddress,
				}
				networkSettings = append(networkSettings, ns)
				if idx == 0 {
					c["ip_address"] = net.IPAddress
				}
			}
		}
		c["network_settings"] = networkSettings

		containers[i] = c
	}

	if err := d.Set("container", containers); err != nil {
		return diag.FromErr(fmt.Errorf("error setting container attribute: %s", err))
	}

	return diags
}

// determineProjectStatus returns the overall project status based on container states.
// Returns "running" if all containers are running, "stopped" if none are running,
// and "partial" if there is a mix of states.
func determineProjectStatus(entries []ComposePSEntry) string {
	if len(entries) == 0 {
		return "stopped"
	}

	allRunning := true
	anyRunning := false

	for _, e := range entries {
		if strings.ToLower(e.State) == "running" {
			anyRunning = true
		} else {
			allRunning = false
		}
	}

	if allRunning {
		return "running"
	}
	if anyRunning {
		return "partial"
	}
	return "stopped"
}
