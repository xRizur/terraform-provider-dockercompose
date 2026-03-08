package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/xRizur/terraform-provider-dockercompose/internal/docker"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// ============================================================
// Terraform Acceptance Tests
//
// These tests require Docker to be running and will create real containers.
// Run with: TF_ACC=1 go test -v -run TestAcc -timeout 10m
// ============================================================

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"dockercompose": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	// Check Docker is available
	client := &docker.DockerClient{Binary: "docker"}
	_, err := client.Version()
	if err != nil {
		t.Skipf("Docker not available, skipping acceptance test: %s", err)
	}
}

// ============================================================
// Stack resource acceptance tests
// ============================================================

func TestAccStackBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-basic"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockercompose_stack.test", "name", "acc-basic"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.test", "compose_yaml"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.test", "compose_file_path"),
					testAccCheckStackRunning("acc-basic"),
					testAccCheckComposeFileExists("dockercompose_stack.test"),
					testAccCheckYAMLContains("dockercompose_stack.test", "image: nginx:alpine"),
					// Container runtime attributes
					resource.TestCheckResourceAttr("dockercompose_stack.test", "container.#", "1"),
					resource.TestCheckResourceAttr("dockercompose_stack.test", "container.0.service", "web"),
					resource.TestCheckResourceAttr("dockercompose_stack.test", "container.0.image", "nginx:alpine"),
					resource.TestCheckResourceAttr("dockercompose_stack.test", "container.0.state", "running"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.test", "container.0.container_id"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.test", "container.0.container_name"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.test", "container.0.ip_address"),
				),
			},
		},
	})
}

func TestAccStackMultipleServices(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-multi"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigMultiService(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockercompose_stack.multi", "name", "acc-multi"),
					testAccCheckStackRunning("acc-multi"),
					testAccCheckYAMLContains("dockercompose_stack.multi", "nginx:alpine"),
					testAccCheckYAMLContains("dockercompose_stack.multi", "redis:7-alpine"),
					// 2 containers sorted alphabetically by service: cache, web
					resource.TestCheckResourceAttr("dockercompose_stack.multi", "container.#", "2"),
					resource.TestCheckResourceAttr("dockercompose_stack.multi", "container.0.service", "cache"),
					resource.TestCheckResourceAttr("dockercompose_stack.multi", "container.0.image", "redis:7-alpine"),
					resource.TestCheckResourceAttr("dockercompose_stack.multi", "container.1.service", "web"),
					resource.TestCheckResourceAttr("dockercompose_stack.multi", "container.1.image", "nginx:alpine"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.multi", "container.0.ip_address"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.multi", "container.1.ip_address"),
				),
			},
		},
	})
}

func TestAccStackWithNetwork(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-net"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigWithNetwork(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-net"),
					testAccCheckYAMLContains("dockercompose_stack.nettest", "networks:"),
					// Container should have network_settings with the custom network
					resource.TestCheckResourceAttr("dockercompose_stack.nettest", "container.#", "1"),
					resource.TestCheckResourceAttr("dockercompose_stack.nettest", "container.0.network_settings.#", "1"),
					resource.TestCheckResourceAttrSet("dockercompose_stack.nettest", "container.0.network_settings.0.ip_address"),
				),
			},
		},
	})
}

func TestAccStackWithVolume(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-vol"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigWithVolume(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-vol"),
					testAccCheckYAMLContains("dockercompose_stack.voltest", "volumes:"),
				),
			},
		},
	})
}

func TestAccStackUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-update"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigUpdate("nginx:alpine"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-update"),
					testAccCheckYAMLContains("dockercompose_stack.updatetest", "nginx:alpine"),
				),
			},
			{
				Config: testAccStackConfigUpdate("nginx:latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-update"),
					testAccCheckYAMLContains("dockercompose_stack.updatetest", "nginx:latest"),
				),
			},
		},
	})
}

func TestAccStackWithEnvironment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-env"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigWithEnv(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-env"),
					testAccCheckYAMLContains("dockercompose_stack.envtest", "APP_ENV: production"),
				),
			},
		},
	})
}

func TestAccStackWithHealthcheck(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-health"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigWithHealthcheck(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-health"),
					testAccCheckYAMLContains("dockercompose_stack.healthtest", "healthcheck:"),
					testAccCheckYAMLContains("dockercompose_stack.healthtest", "interval: 10s"),
				),
			},
		},
	})
}

func TestAccStackWithAllServiceOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-full"),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfigFull(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-full"),
					testAccCheckYAMLContains("dockercompose_stack.fulltest", "hostname: webhost"),
					testAccCheckYAMLContains("dockercompose_stack.fulltest", "stop_signal: SIGTERM"),
					testAccCheckYAMLContains("dockercompose_stack.fulltest", "shm_size: 64m"),
				),
			},
		},
	})
}

// ============================================================
// Project resource acceptance tests
// ============================================================

func TestAccProjectInlineYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-project-inline"),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfigInline(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockercompose_project.inline", "name", "acc-project-inline"),
					resource.TestCheckResourceAttrSet("dockercompose_project.inline", "yaml_sha256"),
					testAccCheckStackRunning("acc-project-inline"),
					// Container attributes also work on dockercompose_project
					resource.TestCheckResourceAttr("dockercompose_project.inline", "container.#", "1"),
					resource.TestCheckResourceAttr("dockercompose_project.inline", "container.0.service", "web"),
					resource.TestCheckResourceAttr("dockercompose_project.inline", "container.0.state", "running"),
					resource.TestCheckResourceAttrSet("dockercompose_project.inline", "container.0.container_id"),
					resource.TestCheckResourceAttrSet("dockercompose_project.inline", "container.0.ip_address"),
				),
			},
		},
	})
}

func TestAccProjectFromFile(t *testing.T) {
	// Create a temporary compose file
	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	content := `services:
  web:
    image: nginx:alpine
    ports:
      - "18082:80"
`
	if err := os.WriteFile(composeFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp compose file: %s", err)
	}

	// Convert backslashes to forward slashes for Terraform HCL
	hclPath := strings.ReplaceAll(composeFile, "\\", "/")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-project-file"),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "dockercompose_project" "fromfile" {
  name         = "acc-project-file"
  compose_file = "%s"
}
`, hclPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-project-file"),
				),
			},
		},
	})
}

// ============================================================
// Regression tests
// ============================================================

func TestAccStackDestroyRemovesContainers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-destroy-test"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "dockercompose_stack" "destroytest" {
  name = "acc-destroy-test"
  service {
    name  = "web"
    image = "nginx:alpine"
  }
}
`,
				Check: testAccCheckStackRunning("acc-destroy-test"),
			},
		},
	})
}

func TestAccStackEmptyPortsList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-no-ports"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "dockercompose_stack" "noports" {
  name = "acc-no-ports"
  service {
    name  = "worker"
    image = "alpine:latest"
    command = ["sleep", "3600"]
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackRunning("acc-no-ports"),
					testAccCheckYAMLNotContains("dockercompose_stack.noports", "ports:"),
				),
			},
		},
	})
}

func TestAccStackIdempotent(t *testing.T) {
	config := `
resource "dockercompose_stack" "idempotent" {
  name = "acc-idempotent"
  service {
    name  = "web"
    image = "nginx:alpine"
    ports = ["18083:80"]
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-idempotent"),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  testAccCheckStackRunning("acc-idempotent"),
			},
			{
				// Apply same config again — should produce no changes
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// ============================================================
// Check functions (test helpers)
// ============================================================

func testAccCheckStackRunning(projectName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := &docker.DockerClient{Binary: "docker"}
		output, err := client.ComposePSServices(projectName, "")
		if err != nil {
			return fmt.Errorf("docker compose ps failed for project %q: %s", projectName, err)
		}
		if strings.TrimSpace(output) == "" {
			return fmt.Errorf("no running services found for project %q", projectName)
		}
		return nil
	}
}

func testAccCheckStackDestroy(projectName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := &docker.DockerClient{Binary: "docker"}
		output, _ := client.ComposePSServices(projectName, "")
		if strings.TrimSpace(output) != "" {
			return fmt.Errorf("stack %q still has running services: %s", projectName, output)
		}
		return nil
	}
}

func testAccCheckComposeFileExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		filePath := rs.Primary.Attributes["compose_file_path"]
		if filePath == "" {
			return fmt.Errorf("compose_file_path is empty")
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("compose file does not exist: %s", filePath)
		}
		return nil
	}
}

func testAccCheckYAMLContains(resourceName, substr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		yaml := rs.Primary.Attributes["compose_yaml"]
		if yaml == "" {
			return fmt.Errorf("compose_yaml is empty")
		}
		if !strings.Contains(yaml, substr) {
			return fmt.Errorf("compose_yaml does not contain %q.\nFull YAML:\n%s", substr, yaml)
		}
		return nil
	}
}

func testAccCheckYAMLNotContains(resourceName, substr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		yaml := rs.Primary.Attributes["compose_yaml"]
		if strings.Contains(yaml, substr) {
			return fmt.Errorf("compose_yaml should NOT contain %q.\nFull YAML:\n%s", substr, yaml)
		}
		return nil
	}
}

// ============================================================
// Test configs
// ============================================================

func testAccStackConfigBasic() string {
	return `
resource "dockercompose_stack" "test" {
  name = "acc-basic"
  service {
    name  = "web"
    image = "nginx:alpine"
    ports = ["18080:80"]
  }
}
`
}

func testAccStackConfigMultiService() string {
	return `
resource "dockercompose_stack" "multi" {
  name = "acc-multi"

  service {
    name  = "web"
    image = "nginx:alpine"
    ports = ["18090:80"]
  }

  service {
    name  = "cache"
    image = "redis:7-alpine"
  }
}
`
}

func testAccStackConfigWithNetwork() string {
	return `
resource "dockercompose_stack" "nettest" {
  name = "acc-net"

  service {
    name     = "web"
    image    = "nginx:alpine"
    networks = ["mynet"]
  }

  network {
    name   = "mynet"
    driver = "bridge"
  }
}
`
}

func testAccStackConfigWithVolume() string {
	return `
resource "dockercompose_stack" "voltest" {
  name = "acc-vol"

  service {
    name    = "web"
    image   = "nginx:alpine"
    volumes = ["testdata:/data"]
  }

  volume {
    name = "testdata"
  }
}
`
}

func testAccStackConfigUpdate(image string) string {
	return fmt.Sprintf(`
resource "dockercompose_stack" "updatetest" {
  name = "acc-update"
  service {
    name  = "web"
    image = "%s"
    ports = ["18091:80"]
  }
}
`, image)
}

func testAccStackConfigWithEnv() string {
	return `
resource "dockercompose_stack" "envtest" {
  name = "acc-env"
  service {
    name  = "web"
    image = "nginx:alpine"
    environment = {
      APP_ENV = "production"
      DEBUG   = "false"
    }
  }
}
`
}

func testAccStackConfigWithHealthcheck() string {
	return `
resource "dockercompose_stack" "healthtest" {
  name = "acc-health"
  service {
    name  = "web"
    image = "nginx:alpine"
    ports = ["18092:80"]

    healthcheck_test     = ["CMD", "curl", "-f", "http://localhost"]
    healthcheck_interval = "10s"
    healthcheck_timeout  = "5s"
    healthcheck_retries  = 3
  }
}
`
}

func testAccStackConfigFull() string {
	return `
resource "dockercompose_stack" "fulltest" {
  name = "acc-full"

  service {
    name       = "web"
    image      = "nginx:alpine"
    restart    = "unless-stopped"
    ports      = ["18093:80"]
    hostname   = "webhost"
    shm_size   = "64m"
    stop_signal = "SIGTERM"
    stop_grace_period = "10s"

    labels = {
      "test.label" = "value"
    }

    environment = {
      TEST_VAR = "hello"
    }
  }
}
`
}

func testAccProjectConfigInline() string {
	return `
resource "dockercompose_project" "inline" {
  name = "acc-project-inline"
  compose_yaml = <<-EOT
    services:
      web:
        image: nginx:alpine
        ports:
          - "18081:80"
  EOT
}
`
}
