package provider

import (
	"encoding/json"
	"sort"
	"strings"

	"terraform-provider-dockercompose/internal/docker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ============================================================
// JSON structs for Docker CLI output
// ============================================================

// ComposePSEntry represents a single container from `docker compose ps --format json`.
type ComposePSEntry struct {
	ID         string      `json:"ID"`
	Name       string      `json:"Name"`
	Service    string      `json:"Service"`
	Image      string      `json:"Image"`
	State      string      `json:"State"`
	Health     string      `json:"Health"`
	Status     string      `json:"Status"`
	ExitCode   int         `json:"ExitCode"`
	Publishers []Publisher `json:"Publishers"`
}

// Publisher represents a port binding from docker compose ps JSON.
type Publisher struct {
	URL           string `json:"URL"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
	Protocol      string `json:"Protocol"`
}

// DockerInspectEntry holds relevant fields from `docker inspect`.
type DockerInspectEntry struct {
	ID              string `json:"Id"`
	NetworkSettings struct {
		Networks map[string]DockerInspectNetwork `json:"Networks"`
	} `json:"NetworkSettings"`
}

// DockerInspectNetwork holds per-network info from docker inspect.
type DockerInspectNetwork struct {
	IPAddress  string `json:"IPAddress"`
	Gateway    string `json:"Gateway"`
	MacAddress string `json:"MacAddress"`
}

// ============================================================
// Terraform schema for the computed "container" block
// ============================================================

// containerSchema returns the Terraform schema for the computed container block
// shared between dockercompose_stack and dockercompose_project.
func containerSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Runtime information about containers in the stack (populated after apply). Sorted by service name.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"service": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Service name as defined in the compose file.",
				},
				"container_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Docker container ID (short, 12 characters).",
				},
				"container_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Docker container name.",
				},
				"image": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Docker image used by the container.",
				},
				"state": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Container state: running, exited, paused, restarting, dead, etc.",
				},
				"health": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Health status: healthy, unhealthy, starting, or empty if no healthcheck.",
				},
				"exit_code": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Container exit code (0 when running).",
				},
				"ip_address": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "IP address on the first attached network (alphabetically).",
				},
				"ports": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Published port mappings.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"ip": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Bound host IP (e.g. '0.0.0.0'). Empty if not published to host.",
							},
							"private_port": {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: "Container-side port number.",
							},
							"public_port": {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: "Host-side published port number (0 if not published).",
							},
							"protocol": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Port protocol: tcp or udp.",
							},
						},
					},
				},
				"network_settings": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Per-network IP assignment details (from docker inspect).",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Docker network name.",
							},
							"ip_address": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Container IP address on this network.",
							},
							"gateway": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Network gateway IP.",
							},
							"mac_address": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Container MAC address on this network.",
							},
						},
					},
				},
			},
		},
	}
}

// ============================================================
// JSON parsing helpers
// ============================================================

// parseComposePSJSON parses the output of `docker compose ps --format json -a`.
// Handles both NDJSON (one JSON object per line) and JSON array formats.
func parseComposePSJSON(raw string) ([]ComposePSEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	// Try JSON array first
	var entries []ComposePSEntry
	if err := json.Unmarshal([]byte(raw), &entries); err == nil {
		return entries, nil
	}

	// Fall back to NDJSON (one JSON object per line)
	lines := strings.Split(raw, "\n")
	entries = make([]ComposePSEntry, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry ComposePSEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// parseDockerInspect parses the JSON array output of `docker inspect`.
func parseDockerInspect(raw string) ([]DockerInspectEntry, error) {
	var entries []DockerInspectEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// ============================================================
// Shared read logic
// ============================================================

// readContainerInfo queries Docker for container runtime details and sets the
// "container" attribute on the ResourceData. Shared between stack and project.
func readContainerInfo(d *schema.ResourceData, client *docker.DockerClient, stackName, composeFilePath string) error {
	// Get container list via docker compose ps --format json
	psJSON, err := client.ComposePSJSON(stackName, composeFilePath)
	if err != nil {
		d.Set("container", []interface{}{})
		return nil
	}

	entries, err := parseComposePSJSON(psJSON)
	if err != nil || len(entries) == 0 {
		d.Set("container", []interface{}{})
		return nil
	}

	// Collect container names for docker inspect (names are unambiguous)
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
			// Sort network names for deterministic output
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
				// Use first network's IP as the top-level ip_address
				if idx == 0 {
					c["ip_address"] = net.IPAddress
				}
			}
		}
		c["network_settings"] = networkSettings

		containers[i] = c
	}

	return d.Set("container", containers)
}
