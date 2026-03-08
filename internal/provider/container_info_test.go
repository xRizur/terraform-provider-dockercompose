package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ============================================================
// Unit Tests for container_info.go
// ============================================================

// --- parseComposePSJSON ---

func TestParseComposePSJSON_Empty(t *testing.T) {
	result, err := parseComposePSJSON("")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestParseComposePSJSON_SingleNDJSON(t *testing.T) {
	input := `{"ID":"abc123","Name":"myapp-web-1","Service":"web","Image":"nginx:alpine","State":"running","Health":"healthy","Status":"Up 10 seconds","ExitCode":0,"Publishers":[{"URL":"0.0.0.0","TargetPort":80,"PublishedPort":8080,"Protocol":"tcp"}]}`

	entries, err := parseComposePSJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.ID != "abc123" {
		t.Errorf("ID = %q, want 'abc123'", e.ID)
	}
	if e.Service != "web" {
		t.Errorf("Service = %q, want 'web'", e.Service)
	}
	if e.State != "running" {
		t.Errorf("State = %q, want 'running'", e.State)
	}
	if e.Health != "healthy" {
		t.Errorf("Health = %q, want 'healthy'", e.Health)
	}
	if len(e.Publishers) != 1 {
		t.Fatalf("Publishers count = %d, want 1", len(e.Publishers))
	}
	if e.Publishers[0].PublishedPort != 8080 {
		t.Errorf("PublishedPort = %d, want 8080", e.Publishers[0].PublishedPort)
	}
}

func TestParseComposePSJSON_MultipleNDJSON(t *testing.T) {
	input := `{"ID":"aaa","Name":"app-web-1","Service":"web","Image":"nginx:alpine","State":"running","Health":"","Status":"Up","ExitCode":0,"Publishers":[]}
{"ID":"bbb","Name":"app-db-1","Service":"db","Image":"postgres:17","State":"running","Health":"healthy","Status":"Up","ExitCode":0,"Publishers":[]}`

	entries, err := parseComposePSJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Service != "web" {
		t.Errorf("entries[0].Service = %q, want 'web'", entries[0].Service)
	}
	if entries[1].Service != "db" {
		t.Errorf("entries[1].Service = %q, want 'db'", entries[1].Service)
	}
}

func TestParseComposePSJSON_JSONArray(t *testing.T) {
	entries := []ComposePSEntry{
		{
			ID:      "aaa",
			Name:    "app-web-1",
			Service: "web",
			Image:   "nginx:alpine",
			State:   "running",
			Health:  "healthy",
		},
		{
			ID:      "bbb",
			Name:    "app-db-1",
			Service: "db",
			Image:   "postgres:17",
			State:   "running",
		},
	}
	input, _ := json.Marshal(entries)

	result, err := parseComposePSJSON(string(input))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].ID != "aaa" {
		t.Errorf("result[0].ID = %q, want 'aaa'", result[0].ID)
	}
	if result[1].ID != "bbb" {
		t.Errorf("result[1].ID = %q, want 'bbb'", result[1].ID)
	}
}

func TestParseComposePSJSON_WithBlankLines(t *testing.T) {
	input := `
{"ID":"aaa","Name":"a","Service":"web","Image":"nginx","State":"running","Health":"","Status":"Up","ExitCode":0,"Publishers":[]}

{"ID":"bbb","Name":"b","Service":"db","Image":"pg","State":"running","Health":"","Status":"Up","ExitCode":0,"Publishers":[]}
`
	entries, err := parseComposePSJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestParseComposePSJSON_ExitedContainer(t *testing.T) {
	input := `{"ID":"xyz","Name":"app-worker-1","Service":"worker","Image":"alpine","State":"exited","Health":"","Status":"Exited (1) 5 minutes ago","ExitCode":1,"Publishers":[]}`

	entries, err := parseComposePSJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].State != "exited" {
		t.Errorf("State = %q, want 'exited'", entries[0].State)
	}
	if entries[0].ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", entries[0].ExitCode)
	}
}

func TestParseComposePSJSON_MultiplePublishers(t *testing.T) {
	input := `{"ID":"abc","Name":"app-web-1","Service":"web","Image":"nginx","State":"running","Health":"","Status":"Up","ExitCode":0,"Publishers":[{"URL":"0.0.0.0","TargetPort":80,"PublishedPort":8080,"Protocol":"tcp"},{"URL":"::","TargetPort":80,"PublishedPort":8080,"Protocol":"tcp"},{"URL":"","TargetPort":443,"PublishedPort":0,"Protocol":"tcp"}]}`

	entries, err := parseComposePSJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	pubs := entries[0].Publishers
	if len(pubs) != 3 {
		t.Fatalf("expected 3 publishers, got %d", len(pubs))
	}
	if pubs[0].URL != "0.0.0.0" {
		t.Errorf("pubs[0].URL = %q, want '0.0.0.0'", pubs[0].URL)
	}
	if pubs[1].URL != "::" {
		t.Errorf("pubs[1].URL = %q, want '::'", pubs[1].URL)
	}
	if pubs[2].PublishedPort != 0 {
		t.Errorf("pubs[2].PublishedPort = %d, want 0", pubs[2].PublishedPort)
	}
}

func TestParseComposePSJSON_InvalidJSON(t *testing.T) {
	input := `not valid json at all`
	_, err := parseComposePSJSON(input)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// --- parseDockerInspect ---

func TestParseDockerInspect_Empty(t *testing.T) {
	result, err := parseDockerInspect("[]")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestParseDockerInspect_SingleContainer(t *testing.T) {
	input := `[{
		"Id": "abc123def456789012345678901234567890123456789012345678901234",
		"NetworkSettings": {
			"Networks": {
				"myapp_backend": {
					"IPAddress": "172.28.0.2",
					"Gateway": "172.28.0.1",
					"MacAddress": "02:42:ac:1c:00:02"
				}
			}
		}
	}]`

	entries, err := parseDockerInspect(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	nets := entries[0].NetworkSettings.Networks
	if len(nets) != 1 {
		t.Fatalf("expected 1 network, got %d", len(nets))
	}

	net, ok := nets["myapp_backend"]
	if !ok {
		t.Fatal("expected 'myapp_backend' network")
	}
	if net.IPAddress != "172.28.0.2" {
		t.Errorf("IPAddress = %q, want '172.28.0.2'", net.IPAddress)
	}
	if net.Gateway != "172.28.0.1" {
		t.Errorf("Gateway = %q, want '172.28.0.1'", net.Gateway)
	}
	if net.MacAddress != "02:42:ac:1c:00:02" {
		t.Errorf("MacAddress = %q, want '02:42:ac:1c:00:02'", net.MacAddress)
	}
}

func TestParseDockerInspect_MultipleNetworks(t *testing.T) {
	input := `[{
		"Id": "abc123def456",
		"NetworkSettings": {
			"Networks": {
				"frontend": {
					"IPAddress": "172.18.0.3",
					"Gateway": "172.18.0.1",
					"MacAddress": "aa:bb:cc:dd:ee:01"
				},
				"backend": {
					"IPAddress": "172.28.0.5",
					"Gateway": "172.28.0.1",
					"MacAddress": "aa:bb:cc:dd:ee:02"
				}
			}
		}
	}]`

	entries, err := parseDockerInspect(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	nets := entries[0].NetworkSettings.Networks
	if len(nets) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(nets))
	}

	if nets["frontend"].IPAddress != "172.18.0.3" {
		t.Errorf("frontend IP = %q, want '172.18.0.3'", nets["frontend"].IPAddress)
	}
	if nets["backend"].IPAddress != "172.28.0.5" {
		t.Errorf("backend IP = %q, want '172.28.0.5'", nets["backend"].IPAddress)
	}
}

func TestParseDockerInspect_InvalidJSON(t *testing.T) {
	_, err := parseDockerInspect("invalid")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseDockerInspect_MultipleContainers(t *testing.T) {
	input := `[
		{"Id": "aaa111", "NetworkSettings": {"Networks": {"net1": {"IPAddress": "10.0.0.1", "Gateway": "10.0.0.254", "MacAddress": "aa:aa:aa:aa:aa:01"}}}},
		{"Id": "bbb222", "NetworkSettings": {"Networks": {"net1": {"IPAddress": "10.0.0.2", "Gateway": "10.0.0.254", "MacAddress": "aa:aa:aa:aa:aa:02"}}}}
	]`

	entries, err := parseDockerInspect(input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].NetworkSettings.Networks["net1"].IPAddress != "10.0.0.1" {
		t.Errorf("first container IP = %q, want '10.0.0.1'", entries[0].NetworkSettings.Networks["net1"].IPAddress)
	}
	if entries[1].NetworkSettings.Networks["net1"].IPAddress != "10.0.0.2" {
		t.Errorf("second container IP = %q, want '10.0.0.2'", entries[1].NetworkSettings.Networks["net1"].IPAddress)
	}
}

// --- containerSchema ---

func TestContainerSchema_Structure(t *testing.T) {
	s := containerSchema()

	if s.Type != 5 { // schema.TypeList = 5
		t.Errorf("expected TypeList, got %d", s.Type)
	}
	if !s.Computed {
		t.Error("expected Computed = true")
	}

	elemResource, ok := s.Elem.(*schema.Resource)
	if !ok {
		t.Fatal("expected Elem to be *schema.Resource")
	}

	expectedFields := []string{
		"service", "container_id", "container_name", "image",
		"state", "health", "exit_code", "ip_address",
		"ports", "network_settings",
	}

	for _, field := range expectedFields {
		if _, ok := elemResource.Schema[field]; !ok {
			t.Errorf("missing field %q in container schema", field)
		}
	}

	// Verify ports sub-schema
	portsSchema := elemResource.Schema["ports"]
	portsResource, ok := portsSchema.Elem.(*schema.Resource)
	if !ok {
		t.Fatal("expected ports Elem to be *schema.Resource")
	}
	for _, f := range []string{"ip", "private_port", "public_port", "protocol"} {
		if _, ok := portsResource.Schema[f]; !ok {
			t.Errorf("missing field %q in ports sub-schema", f)
		}
	}

	// Verify network_settings sub-schema
	nsSchema := elemResource.Schema["network_settings"]
	nsResource, ok := nsSchema.Elem.(*schema.Resource)
	if !ok {
		t.Fatal("expected network_settings Elem to be *schema.Resource")
	}
	for _, f := range []string{"name", "ip_address", "gateway", "mac_address"} {
		if _, ok := nsResource.Schema[f]; !ok {
			t.Errorf("missing field %q in network_settings sub-schema", f)
		}
	}
}
