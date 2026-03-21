package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ============================================================
// Unit Tests for data_source_project.go
// ============================================================

// --- determineProjectStatus ---

func TestDetermineProjectStatus_AllRunning(t *testing.T) {
	entries := []ComposePSEntry{
		{ID: "aaa", Service: "web", State: "running"},
		{ID: "bbb", Service: "db", State: "running"},
	}
	status := determineProjectStatus(entries)
	if status != "running" {
		t.Errorf("expected 'running', got %q", status)
	}
}

func TestDetermineProjectStatus_AllExited(t *testing.T) {
	entries := []ComposePSEntry{
		{ID: "aaa", Service: "web", State: "exited"},
		{ID: "bbb", Service: "db", State: "exited"},
	}
	status := determineProjectStatus(entries)
	if status != "stopped" {
		t.Errorf("expected 'stopped', got %q", status)
	}
}

func TestDetermineProjectStatus_Mixed(t *testing.T) {
	entries := []ComposePSEntry{
		{ID: "aaa", Service: "web", State: "running"},
		{ID: "bbb", Service: "db", State: "exited"},
	}
	status := determineProjectStatus(entries)
	if status != "partial" {
		t.Errorf("expected 'partial', got %q", status)
	}
}

func TestDetermineProjectStatus_Empty(t *testing.T) {
	entries := []ComposePSEntry{}
	status := determineProjectStatus(entries)
	if status != "stopped" {
		t.Errorf("expected 'stopped', got %q", status)
	}
}

func TestDetermineProjectStatus_SingleRunning(t *testing.T) {
	entries := []ComposePSEntry{
		{ID: "aaa", Service: "web", State: "running"},
	}
	status := determineProjectStatus(entries)
	if status != "running" {
		t.Errorf("expected 'running', got %q", status)
	}
}

func TestDetermineProjectStatus_Restarting(t *testing.T) {
	entries := []ComposePSEntry{
		{ID: "aaa", Service: "web", State: "running"},
		{ID: "bbb", Service: "db", State: "restarting"},
	}
	status := determineProjectStatus(entries)
	if status != "partial" {
		t.Errorf("expected 'partial', got %q", status)
	}
}

// --- dataSourceComposeProject schema validation ---

func TestDataSourceProjectSchema(t *testing.T) {
	ds := dataSourceComposeProject()

	// Verify ReadContext function is set
	if ds.ReadContext == nil {
		t.Error("data source should have a ReadContext function")
	}

	// Verify 'name' is required
	nameSchema, ok := ds.Schema["name"]
	if !ok {
		t.Fatal("data source schema missing 'name' field")
	}
	if !nameSchema.Required {
		t.Error("'name' should be Required")
	}
	if nameSchema.Type != schema.TypeString {
		t.Error("'name' should be TypeString")
	}

	// Verify 'status' is computed
	statusSchema, ok := ds.Schema["status"]
	if !ok {
		t.Fatal("data source schema missing 'status' field")
	}
	if !statusSchema.Computed {
		t.Error("'status' should be Computed")
	}
	if statusSchema.Type != schema.TypeString {
		t.Error("'status' should be TypeString")
	}

	// Verify 'container' is computed (reuses containerSchema)
	containerSch, ok := ds.Schema["container"]
	if !ok {
		t.Fatal("data source schema missing 'container' field")
	}
	if !containerSch.Computed {
		t.Error("'container' should be Computed")
	}
	if containerSch.Type != schema.TypeList {
		t.Error("'container' should be TypeList")
	}
}

func TestDataSourceProjectSchema_NoCreateUpdateDelete(t *testing.T) {
	ds := dataSourceComposeProject()

	if ds.CreateContext != nil {
		t.Error("data source should NOT have a CreateContext function")
	}
	if ds.UpdateContext != nil {
		t.Error("data source should NOT have an UpdateContext function")
	}
	if ds.DeleteContext != nil {
		t.Error("data source should NOT have a DeleteContext function")
	}
}
