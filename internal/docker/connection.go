package docker

import (
	"fmt"
	"strings"
)

// ConnectionConfig holds SSH connection parameters for establishing a Docker
// host URL from resource attributes. Only SSH is supported.
type ConnectionConfig struct {
	Host string
	User string
	Port int
}

// DockerHostURL builds an ssh:// Docker host URL from the connection config.
func (c *ConnectionConfig) DockerHostURL() string {
	p := c.Port
	if p == 0 {
		p = 22
	}
	if c.User != "" {
		return fmt.Sprintf("ssh://%s@%s:%d", c.User, c.Host, p)
	}
	return fmt.Sprintf("ssh://%s:%d", c.Host, p)
}

// EffectiveHost resolves the Docker host URL using the priority:
//
//	resource.ssh_connection > resource.host > provider.host
func EffectiveHost(resourceHost string, conn *ConnectionConfig, providerHost string) (string, error) {
	if conn != nil {
		return conn.DockerHostURL(), nil
	}
	if resourceHost != "" {
		return resourceHost, nil
	}
	if providerHost != "" {
		return providerHost, nil
	}
	return "", fmt.Errorf("no Docker host configured: set provider.host, resource.host, or resource.ssh_connection")
}

// ClientForHost returns a new DockerClient with the given host URL, copying
// the binary and project directory settings from the base client.
func ClientForHost(host string, base *DockerClient) *DockerClient {
	return &DockerClient{
		Host:             host,
		Binary:           base.Binary,
		ProjectDirectory: base.ProjectDirectory,
	}
}

// ComposeResourceID builds the Terraform resource ID for a dockercompose
// resource. When a resource-level host is set the ID encodes both so that
// resources on different nodes are distinguishable in state.
func ComposeResourceID(host, projectName string) string {
	if host == "" {
		return projectName
	}
	return host + "/" + projectName
}

// ParseComposeResourceID splits a resource ID back into host and project name.
// The host portion is only present when the ID was built with ComposeResourceID
// and a non-empty host (i.e. it contains "://").
func ParseComposeResourceID(id string) (host, projectName string) {
	idx := strings.LastIndex(id, "/")
	if idx > 0 && strings.Contains(id[:idx], "://") {
		return id[:idx], id[idx+1:]
	}
	return "", id
}
