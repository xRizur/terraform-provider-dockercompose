package docker

import (
	"strings"
	"testing"
)

// ============================================================
// Unit Tests for compose.go: YAML marshaling/unmarshaling
// ============================================================

func TestMarshalMinimalService(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web": {Image: "nginx:latest"},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "services:")
	assertContains(t, yaml, "web:")
	assertContains(t, yaml, "image: nginx:latest")

	// Should NOT contain zero-value optional fields
	assertNotContains(t, yaml, "networks:")
	assertNotContains(t, yaml, "volumes:")
	assertNotContains(t, yaml, "privileged:")
	assertNotContains(t, yaml, "read_only:")
	assertNotContains(t, yaml, "stdin_open:")
}

func TestMarshalFullService(t *testing.T) {
	replicas := 3
	retries := 5
	initTrue := true

	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"api": {
				Image:         "myapp:latest",
				ContainerName: "myapp-api",
				Restart:       "unless-stopped",
				Ports:         []string{"3000:3000", "9090:9090"},
				Expose:        []string{"4000"},
				DependsOn:     []string{"db", "redis"},
				Environment:   map[string]string{"NODE_ENV": "production", "PORT": "3000"},
				EnvFile:       []string{".env"},
				Command:       []string{"node", "server.js"},
				Entrypoint:    []string{"/entrypoint.sh"},
				Volumes:       []string{"./data:/data", "dbvol:/db"},
				Networks:      []string{"frontend", "backend"},
				Labels:        map[string]string{"com.example.team": "backend"},
				Deploy: &DeployConfig{
					Replicas: &replicas,
					Resources: &DeployResources{
						Limits:       &ResourceSpec{Cpus: "0.5", Memory: "512M"},
						Reservations: &ResourceSpec{Cpus: "0.25", Memory: "256M"},
					},
				},
				Healthcheck: &HealthcheckCfg{
					Test:        []string{"CMD", "curl", "-f", "http://localhost:3000/health"},
					Interval:    "30s",
					Timeout:     "10s",
					Retries:     &retries,
					StartPeriod: "40s",
				},
				Logging: &LoggingConfig{
					Driver:  "json-file",
					Options: map[string]string{"max-size": "10m", "max-file": "3"},
				},
				CapAdd:          []string{"NET_ADMIN"},
				CapDrop:         []string{"ALL"},
				SecurityOpt:     []string{"no-new-privileges:true"},
				Privileged:      true,
				ReadOnly:        true,
				Init:            &initTrue,
				User:            "1000:1000",
				DNS:             []string{"8.8.8.8"},
				ExtraHosts:      []string{"host.docker.internal:host-gateway"},
				Hostname:        "api",
				Domainname:      "example.com",
				WorkingDir:      "/app",
				StdinOpen:       true,
				Tty:             true,
				ShmSize:         "256m",
				StopGracePeriod: "30s",
				StopSignal:      "SIGTERM",
				Platform:        "linux/amd64",
				PullPolicy:      "always",
				Runtime:         "runc",
				Tmpfs:           []string{"/tmp"},
				Devices:         []string{"/dev/sda:/dev/xvdc:rwm"},
				Sysctls:         map[string]string{"net.core.somaxconn": "1024"},
				Profiles:        []string{"debug"},
				Pid:             "host",
				Ipc:             "host",
			},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)

	// Verify all major sections are present
	checks := []string{
		"image: myapp:latest",
		"container_name: myapp-api",
		"restart: unless-stopped",
		"- 3000:3000",
		"- 9090:9090",
		"- \"4000\"",
		"- db",
		"- redis",
		"NODE_ENV: production",
		"- .env",
		"- node",
		"- /entrypoint.sh",
		"- ./data:/data",
		"- frontend",
		"com.example.team: backend",
		"replicas: 3",
		"cpus: \"0.5\"",
		"memory: 512M",
		"- CMD",
		"interval: 30s",
		"retries: 5",
		"start_period: 40s",
		"driver: json-file",
		"max-size: 10m",
		"- NET_ADMIN",
		"- ALL",
		"- no-new-privileges:true",
		"privileged: true",
		"read_only: true",
		"init: true",
		"user: 1000:1000",
		"- 8.8.8.8",
		"- host.docker.internal:host-gateway",
		"hostname: api",
		"domainname: example.com",
		"working_dir: /app",
		"stdin_open: true",
		"tty: true",
		"shm_size: 256m",
		"stop_grace_period: 30s",
		"stop_signal: SIGTERM",
		"platform: linux/amd64",
		"pull_policy: always",
		"runtime: runc",
		"- /tmp",
		"- /dev/sda:/dev/xvdc:rwm",
		"net.core.somaxconn: \"1024\"",
		"- debug",
		"pid: host",
		"ipc: host",
	}

	for _, check := range checks {
		assertContains(t, yaml, check)
	}
}

func TestMarshalNetworks(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web": {Image: "nginx:latest"},
		},
		Networks: map[string]*NetworkConfig{
			"backend": {
				Driver:   "bridge",
				Internal: true,
				IPAM: &IPAMConfig{
					Driver: "default",
					Config: []IPAMPool{{
						Subnet:  "172.28.0.0/16",
						Gateway: "172.28.0.1",
					}},
				},
				Labels: map[string]string{"env": "prod"},
			},
			"frontend": {
				Driver: "bridge",
			},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "networks:")
	assertContains(t, yaml, "backend:")
	assertContains(t, yaml, "driver: bridge")
	assertContains(t, yaml, "internal: true")
	assertContains(t, yaml, "subnet: 172.28.0.0/16")
	assertContains(t, yaml, "gateway: 172.28.0.1")
	assertContains(t, yaml, "env: prod")
}

func TestMarshalVolumes(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web": {Image: "nginx:latest"},
		},
		Volumes: map[string]*VolumeConfig{
			"pgdata": {
				Driver:     "local",
				DriverOpts: map[string]string{"type": "nfs"},
				Labels:     map[string]string{"backup": "daily"},
			},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "volumes:")
	assertContains(t, yaml, "pgdata:")
	assertContains(t, yaml, "driver: local")
	assertContains(t, yaml, "type: nfs")
	assertContains(t, yaml, "backup: daily")
}

func TestMarshalConfigsAndSecrets(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web": {Image: "nginx:latest"},
		},
		Configs: map[string]*ConfigEntry{
			"nginx_conf": {File: "./nginx.conf"},
		},
		Secrets: map[string]*SecretEntry{
			"db_password": {File: "./secrets/db_pass.txt"},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "configs:")
	assertContains(t, yaml, "nginx_conf:")
	assertContains(t, yaml, "file: ./nginx.conf")
	assertContains(t, yaml, "secrets:")
	assertContains(t, yaml, "db_password:")
	assertContains(t, yaml, "file: ./secrets/db_pass.txt")
}

func TestUnmarshalComposeFile(t *testing.T) {
	yamlData := `
services:
  web:
    image: nginx:latest
    ports:
    - "8080:80"
    environment:
      APP_ENV: production
  db:
    image: postgres:16
    restart: always
networks:
  backend:
    driver: bridge
volumes:
  pgdata:
    driver: local
`

	cf, err := UnmarshalComposeFile([]byte(yamlData))
	if err != nil {
		t.Fatalf("UnmarshalComposeFile() error = %v", err)
	}

	if len(cf.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cf.Services))
	}

	web := cf.Services["web"]
	if web == nil {
		t.Fatal("service 'web' not found")
	}
	if web.Image != "nginx:latest" {
		t.Errorf("web.Image = %q, want 'nginx:latest'", web.Image)
	}
	if len(web.Ports) != 1 || web.Ports[0] != "8080:80" {
		t.Errorf("web.Ports = %v, want ['8080:80']", web.Ports)
	}
	if web.Environment["APP_ENV"] != "production" {
		t.Errorf("web.Environment[APP_ENV] = %q, want 'production'", web.Environment["APP_ENV"])
	}

	db := cf.Services["db"]
	if db == nil {
		t.Fatal("service 'db' not found")
	}
	if db.Restart != "always" {
		t.Errorf("db.Restart = %q, want 'always'", db.Restart)
	}

	if len(cf.Networks) != 1 {
		t.Errorf("expected 1 network, got %d", len(cf.Networks))
	}
	if cf.Networks["backend"].Driver != "bridge" {
		t.Errorf("network driver = %q, want 'bridge'", cf.Networks["backend"].Driver)
	}

	if len(cf.Volumes) != 1 {
		t.Errorf("expected 1 volume, got %d", len(cf.Volumes))
	}
	if cf.Volumes["pgdata"].Driver != "local" {
		t.Errorf("volume driver = %q, want 'local'", cf.Volumes["pgdata"].Driver)
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	replicas := 2
	retries := 3
	initTrue := true

	original := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"app": {
				Image:       "myapp:v1.0",
				Restart:     "always",
				Ports:       []string{"8080:8080"},
				Environment: map[string]string{"ENV": "prod"},
				Deploy:      &DeployConfig{Replicas: &replicas},
				Healthcheck: &HealthcheckCfg{
					Test:     []string{"CMD", "curl", "localhost"},
					Interval: "10s",
					Retries:  &retries,
				},
				Init:    &initTrue,
				CapAdd:  []string{"NET_ADMIN"},
				CapDrop: []string{"ALL"},
			},
		},
		Networks: map[string]*NetworkConfig{
			"mynet": {Driver: "bridge"},
		},
		Volumes: map[string]*VolumeConfig{
			"myvol": {Driver: "local"},
		},
	}

	yamlBytes, err := MarshalComposeFile(original)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	restored, err := UnmarshalComposeFile(yamlBytes)
	if err != nil {
		t.Fatalf("UnmarshalComposeFile() error = %v", err)
	}

	// Verify key fields survived round-trip
	app := restored.Services["app"]
	if app == nil {
		t.Fatal("service 'app' not found after round-trip")
	}
	if app.Image != "myapp:v1.0" {
		t.Errorf("Image = %q, want 'myapp:v1.0'", app.Image)
	}
	if app.Restart != "always" {
		t.Errorf("Restart = %q, want 'always'", app.Restart)
	}
	if len(app.Ports) != 1 || app.Ports[0] != "8080:8080" {
		t.Errorf("Ports = %v, want ['8080:8080']", app.Ports)
	}
	if app.Environment["ENV"] != "prod" {
		t.Errorf("Environment[ENV] = %q, want 'prod'", app.Environment["ENV"])
	}
	if app.Deploy == nil || app.Deploy.Replicas == nil || *app.Deploy.Replicas != 2 {
		t.Error("Deploy.Replicas not preserved")
	}
	if app.Healthcheck == nil || app.Healthcheck.Interval != "10s" {
		t.Error("Healthcheck not preserved")
	}
	if len(app.CapAdd) != 1 || app.CapAdd[0] != "NET_ADMIN" {
		t.Errorf("CapAdd = %v, want [NET_ADMIN]", app.CapAdd)
	}
}

func TestMarshalOmitsEmptyFields(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"simple": {
				Image: "alpine:latest",
				// All other fields left at zero values
			},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)

	// These should NOT appear for a minimal service
	omitted := []string{
		"container_name:", "restart:", "ports:", "expose:", "depends_on:",
		"environment:", "env_file:", "command:", "entrypoint:", "volumes:",
		"labels:", "deploy:", "healthcheck:", "logging:", "cap_add:", "cap_drop:",
		"security_opt:", "privileged:", "read_only:", "init:", "user:",
		"dns:", "extra_hosts:", "hostname:", "domainname:", "network_mode:",
		"working_dir:", "stdin_open:", "tty:", "shm_size:", "stop_grace_period:",
		"stop_signal:", "platform:", "pull_policy:", "runtime:", "tmpfs:",
		"devices:", "sysctls:", "profiles:", "pid:", "ipc:",
		"mem_limit:", "mem_reservation:", "cpus:",
		"networks:", "configs:", "secrets:",
	}

	for _, field := range omitted {
		assertNotContains(t, yaml, field)
	}
}

func TestMarshalExternalNetwork(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web": {Image: "nginx:latest"},
		},
		Networks: map[string]*NetworkConfig{
			"existing_net": {External: true},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "existing_net:")
	assertContains(t, yaml, "external: true")
}

func TestMarshalExternalVolume(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web": {Image: "nginx:latest"},
		},
		Volumes: map[string]*VolumeConfig{
			"shared_data": {External: true},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "shared_data:")
	assertContains(t, yaml, "external: true")
}

func TestMarshalDeployResourcesOnly(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"worker": {
				Image: "worker:latest",
				Deploy: &DeployConfig{
					Resources: &DeployResources{
						Limits: &ResourceSpec{
							Cpus:   "1.0",
							Memory: "1G",
						},
					},
				},
			},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "deploy:")
	assertContains(t, yaml, "resources:")
	assertContains(t, yaml, "limits:")
	assertContains(t, yaml, "cpus: \"1.0\"")
	assertContains(t, yaml, "memory: 1G")
	// No replicas should appear
	assertNotContains(t, yaml, "replicas:")
}

func TestMarshalHealthcheckDisable(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web": {
				Image: "nginx:latest",
				Healthcheck: &HealthcheckCfg{
					Disable: true,
				},
			},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "healthcheck:")
	assertContains(t, yaml, "disable: true")
}

func TestMarshalMultipleServices(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]*ServiceConfig{
			"web":   {Image: "nginx:latest", Ports: []string{"80:80"}},
			"api":   {Image: "api:latest", Ports: []string{"3000:3000"}},
			"db":    {Image: "postgres:16"},
			"redis": {Image: "redis:7"},
		},
	}

	yamlBytes, err := MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("MarshalComposeFile() error = %v", err)
	}

	yaml := string(yamlBytes)
	assertContains(t, yaml, "web:")
	assertContains(t, yaml, "api:")
	assertContains(t, yaml, "db:")
	assertContains(t, yaml, "redis:")
}

// ============================================================
// Helper assertions
// ============================================================

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected YAML to contain %q, but it didn't.\nFull YAML:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected YAML to NOT contain %q, but it did.\nFull YAML:\n%s", substr, s)
	}
}
