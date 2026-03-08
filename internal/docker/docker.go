package docker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerClient wraps docker compose CLI commands with remote host support.
type DockerClient struct {
	Host             string
	Binary           string
	ProjectDirectory string
}

// Version returns the docker compose version details.
func (c *DockerClient) Version() (string, error) {
	return c.compose("", "", "version")
}

// ComposeUp runs `docker compose up -d --remove-orphans`.
func (c *DockerClient) ComposeUp(projectName, composeFile string) (string, error) {
	return c.compose(projectName, composeFile, "up", "-d", "--remove-orphans")
}

// ComposeDown runs `docker compose down`, optionally removing volumes.
func (c *DockerClient) ComposeDown(projectName, composeFile string, removeVolumes bool) (string, error) {
	if removeVolumes {
		return c.compose(projectName, composeFile, "down", "-v")
	}
	return c.compose(projectName, composeFile, "down")
}

// ComposePSJSON returns container status as JSON.
func (c *DockerClient) ComposePSJSON(projectName, composeFile string) (string, error) {
	return c.compose(projectName, composeFile, "ps", "--format", "json", "-a")
}

// ComposePSServices returns the list of running service names.
func (c *DockerClient) ComposePSServices(projectName, composeFile string) (string, error) {
	return c.compose(projectName, composeFile, "ps", "--services")
}

// ComposeConfig validates and outputs the resolved compose config.
func (c *DockerClient) ComposeConfig(projectName, composeFile string) (string, error) {
	return c.compose(projectName, composeFile, "config")
}

// ProjectDir returns the base directory for a stack's compose files.
func (c *DockerClient) ProjectDir(stackName string) string {
	return filepath.Join(c.ProjectDirectory, stackName)
}

// ComposeFilePath returns the default compose file path for a stack.
func (c *DockerClient) ComposeFilePath(stackName string) string {
	return filepath.Join(c.ProjectDir(stackName), "docker-compose.yml")
}

// DockerInspect runs `docker inspect` on one or more containers and returns the JSON output.
func (c *DockerClient) DockerInspect(containerIDs ...string) (string, error) {
	binary := c.Binary
	if binary == "" {
		binary = "docker"
	}

	args := append([]string{"inspect"}, containerIDs...)
	cmd := exec.Command(binary, args...)

	cmd.Env = os.Environ()
	if c.Host != "" {
		cmd.Env = append(cmd.Env, "DOCKER_HOST="+c.Host)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker inspect: %s\n%s", err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

// compose executes a docker compose command with project isolation and remote host support.
func (c *DockerClient) compose(projectName, composeFile string, args ...string) (string, error) {
	binary := c.Binary
	if binary == "" {
		binary = "docker"
	}

	cmdArgs := []string{"compose"}
	if projectName != "" {
		cmdArgs = append(cmdArgs, "-p", projectName)
	}
	if composeFile != "" {
		cmdArgs = append(cmdArgs, "-f", composeFile)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(binary, cmdArgs...)

	// Inherit environment and set DOCKER_HOST if configured
	cmd.Env = os.Environ()
	if c.Host != "" {
		cmd.Env = append(cmd.Env, "DOCKER_HOST="+c.Host)
	}

	// Set working directory for relative path resolution
	if composeFile != "" {
		cmd.Dir = filepath.Dir(composeFile)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker compose %s: %s\n%s",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}
