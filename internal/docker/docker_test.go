package docker

import (
	"path/filepath"
	"testing"
)

// ============================================================
// Unit Tests for docker.go DockerClient path helpers
// ============================================================

func TestDockerClientProjectDir(t *testing.T) {
	client := &DockerClient{
		ProjectDirectory: "/opt/compose",
	}

	result := client.ProjectDir("myapp")
	expected := filepath.Join("/opt/compose", "myapp")
	if result != expected {
		t.Errorf("ProjectDir('myapp') = %q, want %q", result, expected)
	}
}

func TestDockerClientComposeFilePath(t *testing.T) {
	client := &DockerClient{
		ProjectDirectory: "/opt/compose",
	}

	result := client.ComposeFilePath("myapp")
	expected := filepath.Join("/opt/compose", "myapp", "docker-compose.yml")
	if result != expected {
		t.Errorf("ComposeFilePath('myapp') = %q, want %q", result, expected)
	}
}

func TestDockerClientBinaryDefault(t *testing.T) {
	// Verify that empty binary defaults correctly
	client := &DockerClient{
		Binary: "",
	}

	// We can't easily test compose() without Docker, but we can at least
	// verify ProjectDir and ComposeFilePath work with empty ProjectDirectory
	result := client.ProjectDir("test")
	if result != "test" {
		t.Errorf("ProjectDir with empty base = %q, want 'test'", result)
	}
}

func TestDockerClientProjectDirNested(t *testing.T) {
	client := &DockerClient{
		ProjectDirectory: "/home/user/.terraform-docker-compose",
	}

	tests := []struct {
		stack    string
		expected string
	}{
		{"simple", filepath.Join("/home/user/.terraform-docker-compose", "simple")},
		{"my-app", filepath.Join("/home/user/.terraform-docker-compose", "my-app")},
		{"prod_stack_v2", filepath.Join("/home/user/.terraform-docker-compose", "prod_stack_v2")},
	}

	for _, tt := range tests {
		t.Run(tt.stack, func(t *testing.T) {
			result := client.ProjectDir(tt.stack)
			if result != tt.expected {
				t.Errorf("ProjectDir(%q) = %q, want %q", tt.stack, result, tt.expected)
			}
		})
	}
}
