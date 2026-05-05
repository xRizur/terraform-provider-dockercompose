package provider

import (
	"github.com/xRizur/terraform-provider-dockercompose/internal/docker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// connectionSchema returns the schema for the optional "connection" block
// that can be added to any resource or data source.
func connectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"host"},
		Description:   "SSH connection settings used to build the Docker host URL. Conflicts with host.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type":             {Type: schema.TypeString, Optional: true, Default: "ssh", Description: "Connection type. Only 'ssh' is fully supported."},
				"host":             {Type: schema.TypeString, Required: true, Description: "Remote host address."},
				"user":             {Type: schema.TypeString, Optional: true, Description: "SSH username."},
				"port":             {Type: schema.TypeInt, Optional: true, Default: 22, Description: "SSH port. Defaults to 22."},
				"private_key":      {Type: schema.TypeString, Optional: true, Sensitive: true, Description: "PEM-encoded private key."},
				"private_key_file": {Type: schema.TypeString, Optional: true, Description: "Path to the private key file."},
				"known_hosts_file": {Type: schema.TypeString, Optional: true, Description: "Path to a known_hosts file."},
				"bastion_host":     {Type: schema.TypeString, Optional: true, Description: "Bastion / jump host address."},
				"bastion_user":     {Type: schema.TypeString, Optional: true, Description: "Bastion SSH username."},
				"bastion_port":     {Type: schema.TypeInt, Optional: true, Description: "Bastion SSH port."},
			},
		},
	}
}

// connectionFromResourceData extracts and returns a ConnectionConfig from the
// "connection" list attribute in resource data. Returns nil if not set.
func connectionFromResourceData(d *schema.ResourceData) *docker.ConnectionConfig {
	connList, ok := d.GetOk("ssh_connection")
	if !ok {
		return nil
	}
	conns := connList.([]interface{})
	if len(conns) == 0 {
		return nil
	}
	conn := conns[0].(map[string]interface{})
	cfg := &docker.ConnectionConfig{
		Type: conn["type"].(string),
		Host: conn["host"].(string),
	}
	if v, ok := conn["user"].(string); ok {
		cfg.User = v
	}
	if v, ok := conn["port"].(int); ok {
		cfg.Port = v
	}
	if v, ok := conn["private_key"].(string); ok {
		cfg.PrivateKey = v
	}
	if v, ok := conn["private_key_file"].(string); ok {
		cfg.PrivateKeyFile = v
	}
	if v, ok := conn["known_hosts_file"].(string); ok {
		cfg.KnownHostsFile = v
	}
	if v, ok := conn["bastion_host"].(string); ok {
		cfg.BastionHost = v
	}
	if v, ok := conn["bastion_user"].(string); ok {
		cfg.BastionUser = v
	}
	if v, ok := conn["bastion_port"].(int); ok {
		cfg.BastionPort = v
	}
	return cfg
}

// getStr safely extracts a string value from a map.
func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getBool safely extracts a bool value from a map.
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// getBoolPtr returns a *bool. Returns nil for false/unset (omitted from YAML).
func getBoolPtr(m map[string]interface{}, key string) *bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok && b {
			return &b
		}
	}
	return nil
}

// getIntPtr returns a *int. Returns nil for 0/unset (omitted from YAML).
func getIntPtr(m map[string]interface{}, key string) *int {
	if v, ok := m[key]; ok && v != nil {
		if i, ok := v.(int); ok && i != 0 {
			return &i
		}
	}
	return nil
}

// getStrList safely extracts a string slice. Returns nil if empty (omitted from YAML).
func getStrList(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok && v != nil {
		if raw, ok := v.([]interface{}); ok && len(raw) > 0 {
			result := make([]string, 0, len(raw))
			for _, item := range raw {
				if s, ok := item.(string); ok && s != "" {
					result = append(result, s)
				}
			}
			if len(result) > 0 {
				return result
			}
		}
	}
	return nil
}

// getStrMap safely extracts a string map. Returns nil if empty (omitted from YAML).
func getStrMap(m map[string]interface{}, key string) map[string]string {
	if v, ok := m[key]; ok && v != nil {
		if raw, ok := v.(map[string]interface{}); ok && len(raw) > 0 {
			result := make(map[string]string, len(raw))
			for k, val := range raw {
				if s, ok := val.(string); ok {
					result[k] = s
				}
			}
			if len(result) > 0 {
				return result
			}
		}
	}
	return nil
}
