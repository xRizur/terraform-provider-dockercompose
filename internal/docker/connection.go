package docker

import "fmt"

// ConnectionConfig holds SSH (or other protocol) connection parameters for
// establishing a Docker host URL dynamically from resource attributes.
type ConnectionConfig struct {
	Type           string
	Host           string
	User           string
	Port           int
	PrivateKey     string
	PrivateKeyFile string
	KnownHostsFile string
	BastionHost    string
	BastionUser    string
	BastionPort    int
}

// DockerHostURL builds a Docker host URL from the connection config.
// Only type "ssh" is fully supported; other types fall through to the same format.
func (c *ConnectionConfig) DockerHostURL() string {
	t := c.Type
	if t == "" {
		t = "ssh"
	}
	p := c.Port
	if p == 0 {
		p = 22
	}
	if c.User != "" {
		return fmt.Sprintf("%s://%s@%s:%d", t, c.User, c.Host, p)
	}
	return fmt.Sprintf("%s://%s:%d", t, c.Host, p)
}

// EffectiveHost resolves the Docker host URL using the priority:
//
//	resource.connection > resource.host > provider.host
//
// Returns an error if no host can be determined.
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
	return "", fmt.Errorf("no Docker host configured: set provider.host, resource.host, or resource.connection")
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
