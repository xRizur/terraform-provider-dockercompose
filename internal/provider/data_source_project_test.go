package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ============================================================
// Acceptance Tests for dockercompose_project data source
//
// These tests require Docker to be running and will create real containers.
// Run with: TF_ACC=1 go test -v -run TestAccDataSource -timeout 10m
// ============================================================

// TestAccDataSourceProject_Basic creates a stack via a resource, then reads it
// back through the data source and verifies all computed attributes.
func TestAccDataSourceProject_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-ds-basic"),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceProjectBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					// Data source attributes
					resource.TestCheckResourceAttr("data.dockercompose_project.test", "name", "acc-ds-basic"),
					resource.TestCheckResourceAttr("data.dockercompose_project.test", "status", "running"),
					resource.TestCheckResourceAttr("data.dockercompose_project.test", "container.#", "1"),
					resource.TestCheckResourceAttr("data.dockercompose_project.test", "container.0.service", "web"),
					resource.TestCheckResourceAttr("data.dockercompose_project.test", "container.0.image", "nginx:alpine"),
					resource.TestCheckResourceAttr("data.dockercompose_project.test", "container.0.state", "running"),
					resource.TestCheckResourceAttrSet("data.dockercompose_project.test", "container.0.container_id"),
					resource.TestCheckResourceAttrSet("data.dockercompose_project.test", "container.0.container_name"),
					resource.TestCheckResourceAttrSet("data.dockercompose_project.test", "container.0.ip_address"),
				),
			},
		},
	})
}

// TestAccDataSourceProject_MultiService verifies the data source works with
// multiple services and returns sorted container list.
func TestAccDataSourceProject_MultiService(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckStackDestroy("acc-ds-multi"),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceProjectMultiServiceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.dockercompose_project.multi", "name", "acc-ds-multi"),
					resource.TestCheckResourceAttr("data.dockercompose_project.multi", "status", "running"),
					resource.TestCheckResourceAttr("data.dockercompose_project.multi", "container.#", "2"),
					// Containers sorted alphabetically by service name: cache, web
					resource.TestCheckResourceAttr("data.dockercompose_project.multi", "container.0.service", "cache"),
					resource.TestCheckResourceAttr("data.dockercompose_project.multi", "container.0.image", "redis:7-alpine"),
					resource.TestCheckResourceAttr("data.dockercompose_project.multi", "container.1.service", "web"),
					resource.TestCheckResourceAttr("data.dockercompose_project.multi", "container.1.image", "nginx:alpine"),
					resource.TestCheckResourceAttrSet("data.dockercompose_project.multi", "container.0.ip_address"),
					resource.TestCheckResourceAttrSet("data.dockercompose_project.multi", "container.1.ip_address"),
				),
			},
		},
	})
}

// TestAccDataSourceProject_NotFound verifies the data source returns an error
// when the specified project has no running containers.
func TestAccDataSourceProject_NotFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceProjectNotFoundConfig(),
				ExpectError: testAccExpectErrorRegex("no containers found"),
			},
		},
	})
}

// testAccExpectErrorRegex returns a compiled regexp for ExpectError.
func testAccExpectErrorRegex(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

// ============================================================
// Test configs
// ============================================================

func testAccDataSourceProjectBasicConfig() string {
	return `
resource "dockercompose_stack" "ds_basic" {
  name = "acc-ds-basic"
  service {
    name  = "web"
    image = "nginx:alpine"
    ports = ["18100:80"]
  }
}

data "dockercompose_project" "test" {
  name = dockercompose_stack.ds_basic.name
}
`
}

func testAccDataSourceProjectMultiServiceConfig() string {
	return `
resource "dockercompose_stack" "ds_multi" {
  name = "acc-ds-multi"

  service {
    name  = "web"
    image = "nginx:alpine"
    ports = ["18101:80"]
  }

  service {
    name  = "cache"
    image = "redis:7-alpine"
  }
}

data "dockercompose_project" "multi" {
  name = dockercompose_stack.ds_multi.name
}
`
}

func testAccDataSourceProjectNotFoundConfig() string {
	return `
data "dockercompose_project" "missing" {
  name = "nonexistent-project-that-does-not-exist"
}
`
}
