package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ============================================================
// Unit Tests for provider.go schema validation
// ============================================================

func TestProviderSchema(t *testing.T) {
	p := Provider()

	// Validate provider-level schema
	expectedFields := []string{"host", "docker_binary", "project_directory"}
	for _, field := range expectedFields {
		if _, ok := p.Schema[field]; !ok {
			t.Errorf("provider schema missing field %q", field)
		}
	}
}

func TestProviderSchemaDefaults(t *testing.T) {
	p := Provider()

	// docker_binary should default to "docker"
	if p.Schema["docker_binary"].Default != "docker" {
		t.Errorf("docker_binary default = %v, want 'docker'", p.Schema["docker_binary"].Default)
	}

	// host should be optional
	if p.Schema["host"].Required {
		t.Error("host should be optional")
	}

	// host should read from DOCKER_HOST env var
	if p.Schema["host"].DefaultFunc == nil {
		t.Error("host should have DefaultFunc for DOCKER_HOST env")
	}
}

func TestProviderResources(t *testing.T) {
	p := Provider()

	expectedResources := []string{"dockercompose_stack", "dockercompose_project"}
	for _, name := range expectedResources {
		if _, ok := p.ResourcesMap[name]; !ok {
			t.Errorf("provider missing resource %q", name)
		}
	}
}

func TestProviderHasConfigureFunc(t *testing.T) {
	p := Provider()
	if p.ConfigureFunc == nil {
		t.Error("provider ConfigureFunc should not be nil")
	}
}

// TestProviderInternalValidation uses the SDK's built-in validation
func TestProviderInternalValidation(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider InternalValidate() error: %s", err)
	}
}

// ============================================================
// Unit Tests for Stack resource schema
// ============================================================

func TestStackResourceSchema(t *testing.T) {
	r := resourceComposeStack()

	requiredFields := []string{"name", "service"}
	for _, field := range requiredFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Errorf("stack schema missing required field %q", field)
			continue
		}
		if !s.Required {
			t.Errorf("stack field %q should be Required", field)
		}
	}

	optionalFields := []string{
		"working_dir", "remove_volumes_on_destroy",
		"network", "volume", "config", "secret",
	}
	for _, field := range optionalFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Errorf("stack schema missing optional field %q", field)
			continue
		}
		if s.Required {
			t.Errorf("stack field %q should be Optional", field)
		}
	}

	computedFields := []string{"compose_yaml", "compose_file_path"}
	for _, field := range computedFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Errorf("stack schema missing computed field %q", field)
			continue
		}
		if !s.Computed {
			t.Errorf("stack field %q should be Computed", field)
		}
	}
}

func TestStackNameForceNew(t *testing.T) {
	r := resourceComposeStack()
	if !r.Schema["name"].ForceNew {
		t.Error("stack 'name' should ForceNew")
	}
}

func TestStackHasImporter(t *testing.T) {
	r := resourceComposeStack()
	if r.Importer == nil {
		t.Error("stack should have an Importer")
	}
}

func TestStackServiceSchemaComprehensive(t *testing.T) {
	ss := serviceSchema()

	// All expected service fields
	expectedFields := []string{
		// Core
		"name", "image", "container_name", "restart", "ports", "expose",
		"depends_on", "environment", "env_file", "command", "entrypoint",
		"volumes", "networks", "labels",
		// Deploy
		"replicas", "resource_limits_cpus", "resource_limits_memory",
		"resource_reservations_cpus", "resource_reservations_memory",
		// Healthcheck
		"healthcheck_test", "healthcheck_interval", "healthcheck_timeout",
		"healthcheck_retries", "healthcheck_start_period", "healthcheck_disable",
		// Logging
		"logging_driver", "logging_options",
		// Security
		"cap_add", "cap_drop", "security_opt", "privileged", "read_only",
		"init", "user",
		// Networking
		"dns", "extra_hosts", "hostname", "domainname", "network_mode",
		// Runtime
		"working_dir", "stdin_open", "tty", "shm_size", "stop_grace_period",
		"stop_signal", "platform", "pull_policy", "runtime", "tmpfs",
		"devices", "sysctls", "profiles", "pid", "ipc",
		"mem_limit", "mem_reservation", "cpus",
	}

	for _, field := range expectedFields {
		if _, ok := ss[field]; !ok {
			t.Errorf("service schema missing field %q", field)
		}
	}

	// Verify name and image are required
	if !ss["name"].Required {
		t.Error("service.name should be Required")
	}
	if !ss["image"].Required {
		t.Error("service.image should be Required")
	}

	// Verify list types
	listFields := []string{
		"ports", "expose", "depends_on", "env_file", "command", "entrypoint",
		"volumes", "networks", "healthcheck_test", "cap_add", "cap_drop",
		"security_opt", "dns", "extra_hosts", "tmpfs", "devices", "profiles",
	}
	for _, field := range listFields {
		if ss[field].Type != schema.TypeList {
			t.Errorf("service.%s should be TypeList, got %v", field, ss[field].Type)
		}
	}

	// Verify map types
	mapFields := []string{"environment", "labels", "logging_options", "sysctls"}
	for _, field := range mapFields {
		if ss[field].Type != schema.TypeMap {
			t.Errorf("service.%s should be TypeMap, got %v", field, ss[field].Type)
		}
	}
}

func TestNetworkSchemaComplete(t *testing.T) {
	ns := networkSchema()

	expectedFields := []string{
		"name", "driver", "driver_opts", "external", "internal",
		"attachable", "labels", "ipam_driver", "ipam_subnet", "ipam_gateway",
	}

	for _, field := range expectedFields {
		if _, ok := ns[field]; !ok {
			t.Errorf("network schema missing field %q", field)
		}
	}

	if !ns["name"].Required {
		t.Error("network.name should be Required")
	}
}

func TestVolumeSchemaComplete(t *testing.T) {
	vs := volumeSchema()

	expectedFields := []string{
		"name", "driver", "driver_opts", "external", "labels",
	}

	for _, field := range expectedFields {
		if _, ok := vs[field]; !ok {
			t.Errorf("volume schema missing field %q", field)
		}
	}

	if !vs["name"].Required {
		t.Error("volume.name should be Required")
	}
}

// ============================================================
// Unit Tests for Project resource schema
// ============================================================

func TestProjectResourceSchema(t *testing.T) {
	r := resourceComposeProject()

	if !r.Schema["name"].Required {
		t.Error("project.name should be Required")
	}
	if !r.Schema["name"].ForceNew {
		t.Error("project.name should ForceNew")
	}

	// compose_file and compose_yaml should conflict
	if len(r.Schema["compose_file"].ConflictsWith) == 0 {
		t.Error("compose_file should ConflictsWith compose_yaml")
	}
	if len(r.Schema["compose_yaml"].ConflictsWith) == 0 {
		t.Error("compose_yaml should ConflictsWith compose_file")
	}

	// yaml_sha256 should be computed
	if !r.Schema["yaml_sha256"].Computed {
		t.Error("yaml_sha256 should be Computed")
	}
}
