package provider

import (
	"strings"
	"github.com/xRizur/terraform-provider-dockercompose/internal/docker"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// ============================================================
// Integration tests for buildComposeFile
//
// These test the core logic that converts Terraform ResourceData
// into a ComposeFile struct, then marshals it to YAML.
// No Docker or Terraform CLI required.
// ============================================================

func makeResourceData(t *testing.T, attrs map[string]string) *schema.ResourceData {
	t.Helper()
	sm := schema.InternalMap(resourceComposeStack().Schema)
	state := &terraform.InstanceState{
		Attributes: attrs,
	}
	d, err := sm.Data(state, nil)
	if err != nil {
		t.Fatalf("failed to create ResourceData: %s", err)
	}
	return d
}

func TestBuildComposeFileMinimalService(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":            "test-stack",
		"service.#":       "1",
		"service.0.name":  "web",
		"service.0.image": "nginx:latest",
	})

	cf := buildComposeFile(d)

	if cf == nil {
		t.Fatal("buildComposeFile returned nil")
	}
	if len(cf.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cf.Services))
	}
	svc, ok := cf.Services["web"]
	if !ok {
		t.Fatal("expected service 'web'")
	}
	if svc.Image != "nginx:latest" {
		t.Errorf("expected image nginx:latest, got %s", svc.Image)
	}

	// Marshal and verify YAML
	yaml, err := docker.MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	output := string(yaml)
	if !strings.Contains(output, "image: nginx:latest") {
		t.Errorf("YAML missing 'image: nginx:latest':\n%s", output)
	}
	if !strings.Contains(output, "web:") {
		t.Errorf("YAML missing service name 'web:':\n%s", output)
	}
}

func TestBuildComposeFileServiceWithPorts(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":              "test-ports",
		"service.#":         "1",
		"service.0.name":    "web",
		"service.0.image":   "nginx:latest",
		"service.0.ports.#": "2",
		"service.0.ports.0": "8080:80",
		"service.0.ports.1": "8443:443",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["web"]

	if len(svc.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(svc.Ports))
	}
	if svc.Ports[0] != "8080:80" {
		t.Errorf("expected port 8080:80, got %s", svc.Ports[0])
	}
	if svc.Ports[1] != "8443:443" {
		t.Errorf("expected port 8443:443, got %s", svc.Ports[1])
	}

	yaml, _ := docker.MarshalComposeFile(cf)
	output := string(yaml)
	if !strings.Contains(output, "- 8080:80") {
		t.Errorf("YAML missing port 8080:80:\n%s", output)
	}
}

func TestBuildComposeFileServiceWithEnvironment(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                          "test-env",
		"service.#":                     "1",
		"service.0.name":                "app",
		"service.0.image":               "myapp:latest",
		"service.0.environment.%":       "2",
		"service.0.environment.DB_HOST": "postgres",
		"service.0.environment.DB_PORT": "5432",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if svc.Environment == nil {
		t.Fatal("expected environment map")
	}
	if svc.Environment["DB_HOST"] != "postgres" {
		t.Errorf("expected DB_HOST=postgres, got %s", svc.Environment["DB_HOST"])
	}
	if svc.Environment["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432, got %s", svc.Environment["DB_PORT"])
	}

	yaml, _ := docker.MarshalComposeFile(cf)
	output := string(yaml)
	if !strings.Contains(output, "DB_HOST: postgres") {
		t.Errorf("YAML missing DB_HOST:\n%s", output)
	}
}

func TestBuildComposeFileServiceWithVolumes(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                "test-vols",
		"service.#":           "1",
		"service.0.name":      "db",
		"service.0.image":     "postgres:15",
		"service.0.volumes.#": "2",
		"service.0.volumes.0": "pgdata:/var/lib/postgresql/data",
		"service.0.volumes.1": "./init.sql:/docker-entrypoint-initdb.d/init.sql",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["db"]

	if len(svc.Volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(svc.Volumes))
	}
	if svc.Volumes[0] != "pgdata:/var/lib/postgresql/data" {
		t.Errorf("expected volume mapping, got %s", svc.Volumes[0])
	}
}

func TestBuildComposeFileServiceWithDependsOn(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                   "test-deps",
		"service.#":              "2",
		"service.0.name":         "db",
		"service.0.image":        "postgres:15",
		"service.1.name":         "web",
		"service.1.image":        "myapp:latest",
		"service.1.depends_on.#": "1",
		"service.1.depends_on.0": "db",
	})

	cf := buildComposeFile(d)
	web := cf.Services["web"]

	if len(web.DependsOn) != 1 || web.DependsOn[0] != "db" {
		t.Errorf("expected depends_on [db], got %v", web.DependsOn)
	}
}

func TestBuildComposeFileServiceWithRestart(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":              "test-restart",
		"service.#":         "1",
		"service.0.name":    "web",
		"service.0.image":   "nginx:latest",
		"service.0.restart": "unless-stopped",
	})

	cf := buildComposeFile(d)
	if cf.Services["web"].Restart != "unless-stopped" {
		t.Errorf("expected restart 'unless-stopped', got '%s'", cf.Services["web"].Restart)
	}

	yaml, _ := docker.MarshalComposeFile(cf)
	if !strings.Contains(string(yaml), "restart: unless-stopped") {
		t.Error("YAML missing restart policy")
	}
}

func TestBuildComposeFileServiceWithContainerName(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                     "test-cname",
		"service.#":                "1",
		"service.0.name":           "web",
		"service.0.image":          "nginx:latest",
		"service.0.container_name": "my-nginx-container",
	})

	cf := buildComposeFile(d)
	if cf.Services["web"].ContainerName != "my-nginx-container" {
		t.Errorf("expected container_name, got '%s'", cf.Services["web"].ContainerName)
	}
}

func TestBuildComposeFileServiceWithCommand(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                "test-cmd",
		"service.#":           "1",
		"service.0.name":      "app",
		"service.0.image":     "python:3",
		"service.0.command.#": "3",
		"service.0.command.0": "python",
		"service.0.command.1": "-m",
		"service.0.command.2": "http.server",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if len(svc.Command) != 3 {
		t.Fatalf("expected 3 command parts, got %d", len(svc.Command))
	}
	if svc.Command[0] != "python" {
		t.Errorf("expected command[0]='python', got '%s'", svc.Command[0])
	}
}

func TestBuildComposeFileServiceWithDeploy(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                                   "test-deploy",
		"service.#":                              "1",
		"service.0.name":                         "web",
		"service.0.image":                        "nginx:latest",
		"service.0.replicas":                     "3",
		"service.0.resource_limits_cpus":         "0.5",
		"service.0.resource_limits_memory":       "512M",
		"service.0.resource_reservations_cpus":   "0.25",
		"service.0.resource_reservations_memory": "256M",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["web"]

	if svc.Deploy == nil {
		t.Fatal("expected deploy config")
	}
	if svc.Deploy.Replicas == nil || *svc.Deploy.Replicas != 3 {
		t.Error("expected replicas=3")
	}
	if svc.Deploy.Resources == nil {
		t.Fatal("expected deploy resources")
	}
	if svc.Deploy.Resources.Limits.Cpus != "0.5" {
		t.Errorf("expected limits cpus=0.5, got %s", svc.Deploy.Resources.Limits.Cpus)
	}
	if svc.Deploy.Resources.Limits.Memory != "512M" {
		t.Errorf("expected limits memory=512M, got %s", svc.Deploy.Resources.Limits.Memory)
	}
	if svc.Deploy.Resources.Reservations.Cpus != "0.25" {
		t.Errorf("expected reservations cpus=0.25, got %s", svc.Deploy.Resources.Reservations.Cpus)
	}
	if svc.Deploy.Resources.Reservations.Memory != "256M" {
		t.Errorf("expected reservations memory=256M, got %s", svc.Deploy.Resources.Reservations.Memory)
	}

	yaml, _ := docker.MarshalComposeFile(cf)
	output := string(yaml)
	if !strings.Contains(output, "replicas: 3") {
		t.Errorf("YAML missing replicas:\n%s", output)
	}
	if !strings.Contains(output, "cpus: \"0.5\"") {
		t.Errorf("YAML missing cpus limit:\n%s", output)
	}
}

func TestBuildComposeFileServiceWithHealthcheck(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                               "test-hc",
		"service.#":                          "1",
		"service.0.name":                     "web",
		"service.0.image":                    "nginx:latest",
		"service.0.healthcheck_test.#":       "3",
		"service.0.healthcheck_test.0":       "CMD",
		"service.0.healthcheck_test.1":       "curl",
		"service.0.healthcheck_test.2":       "-f http://localhost",
		"service.0.healthcheck_interval":     "30s",
		"service.0.healthcheck_timeout":      "10s",
		"service.0.healthcheck_retries":      "3",
		"service.0.healthcheck_start_period": "5s",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["web"]

	if svc.Healthcheck == nil {
		t.Fatal("expected healthcheck config")
	}
	if len(svc.Healthcheck.Test) != 3 {
		t.Fatalf("expected 3 test parts, got %d", len(svc.Healthcheck.Test))
	}
	if svc.Healthcheck.Test[0] != "CMD" {
		t.Errorf("expected test[0]='CMD', got '%s'", svc.Healthcheck.Test[0])
	}
	if svc.Healthcheck.Interval != "30s" {
		t.Errorf("expected interval=30s, got %s", svc.Healthcheck.Interval)
	}
	if svc.Healthcheck.Timeout != "10s" {
		t.Errorf("expected timeout=10s, got %s", svc.Healthcheck.Timeout)
	}
	if svc.Healthcheck.Retries == nil || *svc.Healthcheck.Retries != 3 {
		t.Error("expected retries=3")
	}
	if svc.Healthcheck.StartPeriod != "5s" {
		t.Errorf("expected start_period=5s, got %s", svc.Healthcheck.StartPeriod)
	}
}

func TestBuildComposeFileServiceWithLogging(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                               "test-log",
		"service.#":                          "1",
		"service.0.name":                     "web",
		"service.0.image":                    "nginx:latest",
		"service.0.logging_driver":           "json-file",
		"service.0.logging_options.%":        "2",
		"service.0.logging_options.max-size": "10m",
		"service.0.logging_options.max-file": "3",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["web"]

	if svc.Logging == nil {
		t.Fatal("expected logging config")
	}
	if svc.Logging.Driver != "json-file" {
		t.Errorf("expected driver=json-file, got %s", svc.Logging.Driver)
	}
	if svc.Logging.Options["max-size"] != "10m" {
		t.Errorf("expected max-size=10m, got %s", svc.Logging.Options["max-size"])
	}
}

func TestBuildComposeFileServiceWithSecurityOptions(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                     "test-security",
		"service.#":                "1",
		"service.0.name":           "app",
		"service.0.image":          "myapp:latest",
		"service.0.cap_add.#":      "2",
		"service.0.cap_add.0":      "NET_ADMIN",
		"service.0.cap_add.1":      "SYS_PTRACE",
		"service.0.cap_drop.#":     "1",
		"service.0.cap_drop.0":     "ALL",
		"service.0.security_opt.#": "1",
		"service.0.security_opt.0": "no-new-privileges:true",
		"service.0.privileged":     "true",
		"service.0.read_only":      "true",
		"service.0.user":           "1000:1000",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if len(svc.CapAdd) != 2 || svc.CapAdd[0] != "NET_ADMIN" {
		t.Errorf("expected cap_add [NET_ADMIN, SYS_PTRACE], got %v", svc.CapAdd)
	}
	if len(svc.CapDrop) != 1 || svc.CapDrop[0] != "ALL" {
		t.Errorf("expected cap_drop [ALL], got %v", svc.CapDrop)
	}
	if len(svc.SecurityOpt) != 1 {
		t.Errorf("expected security_opt, got %v", svc.SecurityOpt)
	}
	if !svc.Privileged {
		t.Error("expected privileged=true")
	}
	if !svc.ReadOnly {
		t.Error("expected read_only=true")
	}
	if svc.User != "1000:1000" {
		t.Errorf("expected user=1000:1000, got %s", svc.User)
	}
}

func TestBuildComposeFileServiceWithNetworkingOptions(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                    "test-net-opts",
		"service.#":               "1",
		"service.0.name":          "app",
		"service.0.image":         "myapp:latest",
		"service.0.dns.#":         "2",
		"service.0.dns.0":         "8.8.8.8",
		"service.0.dns.1":         "1.1.1.1",
		"service.0.extra_hosts.#": "1",
		"service.0.extra_hosts.0": "myhost:127.0.0.1",
		"service.0.hostname":      "myapp-host",
		"service.0.domainname":    "example.com",
		"service.0.network_mode":  "host",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if len(svc.DNS) != 2 || svc.DNS[0] != "8.8.8.8" {
		t.Errorf("expected dns [8.8.8.8, 1.1.1.1], got %v", svc.DNS)
	}
	if len(svc.ExtraHosts) != 1 || svc.ExtraHosts[0] != "myhost:127.0.0.1" {
		t.Errorf("expected extra_hosts, got %v", svc.ExtraHosts)
	}
	if svc.Hostname != "myapp-host" {
		t.Errorf("expected hostname=myapp-host, got %s", svc.Hostname)
	}
	if svc.Domainname != "example.com" {
		t.Errorf("expected domainname=example.com, got %s", svc.Domainname)
	}
	if svc.NetworkMode != "host" {
		t.Errorf("expected network_mode=host, got %s", svc.NetworkMode)
	}
}

func TestBuildComposeFileServiceWithRuntimeOptions(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                        "test-runtime",
		"service.#":                   "1",
		"service.0.name":              "app",
		"service.0.image":             "myapp:latest",
		"service.0.working_dir":       "/app",
		"service.0.stdin_open":        "true",
		"service.0.tty":               "true",
		"service.0.shm_size":          "256m",
		"service.0.stop_grace_period": "30s",
		"service.0.stop_signal":       "SIGTERM",
		"service.0.platform":          "linux/amd64",
		"service.0.pull_policy":       "always",
		"service.0.runtime":           "nvidia",
		"service.0.pid":               "host",
		"service.0.ipc":               "host",
		"service.0.tmpfs.#":           "1",
		"service.0.tmpfs.0":           "/tmp",
		"service.0.devices.#":         "1",
		"service.0.devices.0":         "/dev/sda:/dev/xvdc:rwm",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if svc.WorkingDir != "/app" {
		t.Errorf("expected working_dir=/app, got %s", svc.WorkingDir)
	}
	if !svc.StdinOpen {
		t.Error("expected stdin_open=true")
	}
	if !svc.Tty {
		t.Error("expected tty=true")
	}
	if svc.ShmSize != "256m" {
		t.Errorf("expected shm_size=256m, got %s", svc.ShmSize)
	}
	if svc.StopGracePeriod != "30s" {
		t.Errorf("expected stop_grace_period=30s, got %s", svc.StopGracePeriod)
	}
	if svc.StopSignal != "SIGTERM" {
		t.Errorf("expected stop_signal=SIGTERM, got %s", svc.StopSignal)
	}
	if svc.Platform != "linux/amd64" {
		t.Errorf("expected platform=linux/amd64, got %s", svc.Platform)
	}
	if svc.PullPolicy != "always" {
		t.Errorf("expected pull_policy=always, got %s", svc.PullPolicy)
	}
	if svc.Runtime != "nvidia" {
		t.Errorf("expected runtime=nvidia, got %s", svc.Runtime)
	}
	if svc.Pid != "host" {
		t.Errorf("expected pid=host, got %s", svc.Pid)
	}
	if svc.Ipc != "host" {
		t.Errorf("expected ipc=host, got %s", svc.Ipc)
	}
	if len(svc.Tmpfs) != 1 || svc.Tmpfs[0] != "/tmp" {
		t.Errorf("expected tmpfs [/tmp], got %v", svc.Tmpfs)
	}
	if len(svc.Devices) != 1 || svc.Devices[0] != "/dev/sda:/dev/xvdc:rwm" {
		t.Errorf("expected device mapping, got %v", svc.Devices)
	}
}

func TestBuildComposeFileServiceWithLabelsAndSysctls(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":               "test-labels",
		"service.#":          "1",
		"service.0.name":     "app",
		"service.0.image":    "myapp:latest",
		"service.0.labels.%": "2",
		"service.0.labels.com.example.description": "My App",
		"service.0.labels.traefik.enable":          "true",
		"service.0.sysctls.%":                      "1",
		"service.0.sysctls.net.core.somaxconn":     "1024",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if svc.Labels == nil || svc.Labels["com.example.description"] != "My App" {
		t.Errorf("expected labels, got %v", svc.Labels)
	}
	if svc.Sysctls == nil || svc.Sysctls["net.core.somaxconn"] != "1024" {
		t.Errorf("expected sysctls, got %v", svc.Sysctls)
	}
}

func TestBuildComposeFileServiceWithLegacyResources(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                      "test-legacy",
		"service.#":                 "1",
		"service.0.name":            "app",
		"service.0.image":           "myapp:latest",
		"service.0.mem_limit":       "512m",
		"service.0.mem_reservation": "256m",
		"service.0.cpus":            "0.5",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if svc.MemLimit != "512m" {
		t.Errorf("expected mem_limit=512m, got %s", svc.MemLimit)
	}
	if svc.MemReservation != "256m" {
		t.Errorf("expected mem_reservation=256m, got %s", svc.MemReservation)
	}
	if svc.Cpus != "0.5" {
		t.Errorf("expected cpus=0.5, got %s", svc.Cpus)
	}
}

func TestBuildComposeFileWithNetwork(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":             "test-net",
		"service.#":        "1",
		"service.0.name":   "web",
		"service.0.image":  "nginx:latest",
		"network.#":        "1",
		"network.0.name":   "frontend",
		"network.0.driver": "bridge",
	})

	cf := buildComposeFile(d)

	if cf.Networks == nil {
		t.Fatal("expected networks")
	}
	net, ok := cf.Networks["frontend"]
	if !ok {
		t.Fatal("expected network 'frontend'")
	}
	if net.Driver != "bridge" {
		t.Errorf("expected driver=bridge, got %s", net.Driver)
	}

	yaml, _ := docker.MarshalComposeFile(cf)
	output := string(yaml)
	if !strings.Contains(output, "frontend:") {
		t.Errorf("YAML missing network name:\n%s", output)
	}
	if !strings.Contains(output, "driver: bridge") {
		t.Errorf("YAML missing network driver:\n%s", output)
	}
}

func TestBuildComposeFileWithNetworkIPAM(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                   "test-ipam",
		"service.#":              "1",
		"service.0.name":         "web",
		"service.0.image":        "nginx:latest",
		"network.#":              "1",
		"network.0.name":         "backend",
		"network.0.driver":       "bridge",
		"network.0.ipam_driver":  "default",
		"network.0.ipam_subnet":  "172.28.0.0/16",
		"network.0.ipam_gateway": "172.28.0.1",
	})

	cf := buildComposeFile(d)
	net := cf.Networks["backend"]

	if net.IPAM == nil {
		t.Fatal("expected IPAM config")
	}
	if net.IPAM.Driver != "default" {
		t.Errorf("expected IPAM driver=default, got %s", net.IPAM.Driver)
	}
	if len(net.IPAM.Config) != 1 {
		t.Fatalf("expected 1 IPAM config, got %d", len(net.IPAM.Config))
	}
	if net.IPAM.Config[0].Subnet != "172.28.0.0/16" {
		t.Errorf("expected subnet, got %s", net.IPAM.Config[0].Subnet)
	}
	if net.IPAM.Config[0].Gateway != "172.28.0.1" {
		t.Errorf("expected gateway, got %s", net.IPAM.Config[0].Gateway)
	}
}

func TestBuildComposeFileWithExternalNetwork(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":               "test-ext-net",
		"service.#":          "1",
		"service.0.name":     "web",
		"service.0.image":    "nginx:latest",
		"network.#":          "1",
		"network.0.name":     "existing",
		"network.0.external": "true",
	})

	cf := buildComposeFile(d)
	net := cf.Networks["existing"]

	if !net.External {
		t.Error("expected external=true")
	}
}

func TestBuildComposeFileWithVolume(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":            "test-vol",
		"service.#":       "1",
		"service.0.name":  "db",
		"service.0.image": "postgres:15",
		"volume.#":        "1",
		"volume.0.name":   "pgdata",
		"volume.0.driver": "local",
	})

	cf := buildComposeFile(d)

	if cf.Volumes == nil {
		t.Fatal("expected volumes")
	}
	vol, ok := cf.Volumes["pgdata"]
	if !ok {
		t.Fatal("expected volume 'pgdata'")
	}
	if vol.Driver != "local" {
		t.Errorf("expected driver=local, got %s", vol.Driver)
	}
}

func TestBuildComposeFileWithVolumeDriverOpts(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                        "test-vol-opts",
		"service.#":                   "1",
		"service.0.name":              "app",
		"service.0.image":             "myapp:latest",
		"volume.#":                    "1",
		"volume.0.name":               "nfs-data",
		"volume.0.driver":             "local",
		"volume.0.driver_opts.%":      "3",
		"volume.0.driver_opts.type":   "nfs",
		"volume.0.driver_opts.o":      "addr=10.0.0.1,rw",
		"volume.0.driver_opts.device": ":/data",
	})

	cf := buildComposeFile(d)
	vol := cf.Volumes["nfs-data"]

	if vol.DriverOpts == nil {
		t.Fatal("expected driver_opts")
	}
	if vol.DriverOpts["type"] != "nfs" {
		t.Errorf("expected type=nfs, got %s", vol.DriverOpts["type"])
	}
}

func TestBuildComposeFileWithConfig(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":            "test-cfg",
		"service.#":       "1",
		"service.0.name":  "web",
		"service.0.image": "nginx:latest",
		"config.#":        "1",
		"config.0.name":   "nginx_conf",
		"config.0.file":   "./nginx.conf",
	})

	cf := buildComposeFile(d)

	if cf.Configs == nil {
		t.Fatal("expected configs")
	}
	cfg, ok := cf.Configs["nginx_conf"]
	if !ok {
		t.Fatal("expected config 'nginx_conf'")
	}
	if cfg.File != "./nginx.conf" {
		t.Errorf("expected file=./nginx.conf, got %s", cfg.File)
	}
}

func TestBuildComposeFileWithSecret(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":            "test-sec",
		"service.#":       "1",
		"service.0.name":  "web",
		"service.0.image": "nginx:latest",
		"secret.#":        "1",
		"secret.0.name":   "db_password",
		"secret.0.file":   "./secrets/db_pass.txt",
	})

	cf := buildComposeFile(d)

	if cf.Secrets == nil {
		t.Fatal("expected secrets")
	}
	sec, ok := cf.Secrets["db_password"]
	if !ok {
		t.Fatal("expected secret 'db_password'")
	}
	if sec.File != "./secrets/db_pass.txt" {
		t.Errorf("expected file, got %s", sec.File)
	}
}

func TestBuildComposeFileMultipleServices(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                          "test-multi",
		"service.#":                     "3",
		"service.0.name":                "web",
		"service.0.image":               "nginx:latest",
		"service.0.ports.#":             "1",
		"service.0.ports.0":             "8080:80",
		"service.0.depends_on.#":        "1",
		"service.0.depends_on.0":        "api",
		"service.1.name":                "api",
		"service.1.image":               "myapi:latest",
		"service.1.ports.#":             "1",
		"service.1.ports.0":             "3000:3000",
		"service.1.depends_on.#":        "1",
		"service.1.depends_on.0":        "db",
		"service.1.environment.%":       "1",
		"service.1.environment.DB_HOST": "db",
		"service.2.name":                "db",
		"service.2.image":               "postgres:15",
		"service.2.volumes.#":           "1",
		"service.2.volumes.0":           "pgdata:/var/lib/postgresql/data",
	})

	cf := buildComposeFile(d)

	if len(cf.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(cf.Services))
	}

	// Verify web
	web := cf.Services["web"]
	if web.Image != "nginx:latest" {
		t.Errorf("web: wrong image")
	}
	if len(web.DependsOn) != 1 || web.DependsOn[0] != "api" {
		t.Errorf("web: wrong depends_on")
	}

	// Verify api
	api := cf.Services["api"]
	if api.Environment["DB_HOST"] != "db" {
		t.Errorf("api: wrong environment")
	}

	// Verify db
	db := cf.Services["db"]
	if len(db.Volumes) != 1 {
		t.Errorf("db: wrong volumes")
	}

	// Full YAML round trip
	yaml, err := docker.MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	output := string(yaml)
	if !strings.Contains(output, "web:") {
		t.Error("YAML missing 'web:'")
	}
	if !strings.Contains(output, "api:") {
		t.Error("YAML missing 'api:'")
	}
	if !strings.Contains(output, "db:") {
		t.Error("YAML missing 'db:'")
	}
}

func TestBuildComposeFileFullStack(t *testing.T) {
	// Full-featured stack: services, networks with IPAM, volumes, configs, secrets
	d := makeResourceData(t, map[string]string{
		"name": "full-stack",
		// Service: web with all bells and whistles
		"service.#":                       "2",
		"service.0.name":                  "web",
		"service.0.image":                 "nginx:latest",
		"service.0.container_name":        "my-web",
		"service.0.restart":               "always",
		"service.0.ports.#":               "2",
		"service.0.ports.0":               "80:80",
		"service.0.ports.1":               "443:443",
		"service.0.networks.#":            "1",
		"service.0.networks.0":            "frontend",
		"service.0.depends_on.#":          "1",
		"service.0.depends_on.0":          "api",
		"service.0.healthcheck_test.#":    "3",
		"service.0.healthcheck_test.0":    "CMD",
		"service.0.healthcheck_test.1":    "curl",
		"service.0.healthcheck_test.2":    "-f http://localhost",
		"service.0.healthcheck_interval":  "30s",
		"service.0.healthcheck_timeout":   "10s",
		"service.0.healthcheck_retries":   "5",
		"service.0.labels.%":              "1",
		"service.0.labels.traefik.enable": "true",
		// Service: api
		"service.1.name":                     "api",
		"service.1.image":                    "myapi:v2",
		"service.1.restart":                  "on-failure",
		"service.1.environment.%":            "2",
		"service.1.environment.NODE_ENV":     "production",
		"service.1.environment.PORT":         "3000",
		"service.1.networks.#":               "2",
		"service.1.networks.0":               "frontend",
		"service.1.networks.1":               "backend",
		"service.1.replicas":                 "2",
		"service.1.resource_limits_cpus":     "1.0",
		"service.1.resource_limits_memory":   "1G",
		"service.1.logging_driver":           "json-file",
		"service.1.logging_options.%":        "1",
		"service.1.logging_options.max-size": "10m",
		// Networks
		"network.#":              "2",
		"network.0.name":         "frontend",
		"network.0.driver":       "bridge",
		"network.1.name":         "backend",
		"network.1.driver":       "bridge",
		"network.1.internal":     "true",
		"network.1.ipam_driver":  "default",
		"network.1.ipam_subnet":  "172.28.0.0/16",
		"network.1.ipam_gateway": "172.28.0.1",
		// Volumes
		"volume.#":        "1",
		"volume.0.name":   "app-data",
		"volume.0.driver": "local",
		// Configs
		"config.#":      "1",
		"config.0.name": "nginx_conf",
		"config.0.file": "./nginx.conf",
		// Secrets
		"secret.#":      "1",
		"secret.0.name": "api_key",
		"secret.0.file": "./secrets/api_key.txt",
	})

	cf := buildComposeFile(d)

	// Verify structure
	if len(cf.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cf.Services))
	}
	if len(cf.Networks) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(cf.Networks))
	}
	if len(cf.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(cf.Volumes))
	}
	if len(cf.Configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(cf.Configs))
	}
	if len(cf.Secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(cf.Secrets))
	}

	// Verify web service
	web := cf.Services["web"]
	if web.ContainerName != "my-web" {
		t.Error("web: wrong container_name")
	}
	if web.Restart != "always" {
		t.Error("web: wrong restart")
	}
	if len(web.Ports) != 2 {
		t.Error("web: wrong ports count")
	}
	if web.Healthcheck == nil {
		t.Error("web: missing healthcheck")
	}

	// Verify api service
	api := cf.Services["api"]
	if api.Deploy == nil || api.Deploy.Replicas == nil || *api.Deploy.Replicas != 2 {
		t.Error("api: wrong replicas")
	}
	if api.Logging == nil || api.Logging.Driver != "json-file" {
		t.Error("api: wrong logging")
	}

	// Verify backend network
	backend := cf.Networks["backend"]
	if !backend.Internal {
		t.Error("backend: expected internal=true")
	}
	if backend.IPAM == nil {
		t.Error("backend: missing IPAM")
	}

	// Full YAML output test
	yaml, err := docker.MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	output := string(yaml)

	// Verify YAML has expected sections
	requiredParts := []string{
		"services:",
		"networks:",
		"volumes:",
		"configs:",
		"secrets:",
		"web:",
		"api:",
		"frontend:",
		"backend:",
		"app-data:",
		"nginx_conf:",
		"api_key:",
		"image: nginx:latest",
		"container_name: my-web",
		"restart: always",
		"- 80:80",
		"- 443:443",
		"replicas: 2",
		"driver: bridge",
		"internal: true",
		"subnet: 172.28.0.0/16",
		"gateway: 172.28.0.1",
	}

	for _, part := range requiredParts {
		if !strings.Contains(output, part) {
			t.Errorf("YAML missing '%s':\n%s", part, output)
		}
	}

	// Verify YAML can be round-tripped
	cf2, err := docker.UnmarshalComposeFile(yaml)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if len(cf2.Services) != 2 {
		t.Errorf("round-trip: expected 2 services, got %d", len(cf2.Services))
	}
	if cf2.Services["web"].Image != "nginx:latest" {
		t.Error("round-trip: web image mismatch")
	}
	if cf2.Networks["backend"].IPAM.Config[0].Subnet != "172.28.0.0/16" {
		t.Error("round-trip: backend IPAM subnet mismatch")
	}
}

func TestBuildComposeFileEmptyOptionalFields(t *testing.T) {
	// Verify that a minimal service doesn't produce extraneous YAML fields
	d := makeResourceData(t, map[string]string{
		"name":            "test-min",
		"service.#":       "1",
		"service.0.name":  "web",
		"service.0.image": "nginx:latest",
	})

	cf := buildComposeFile(d)
	yaml, _ := docker.MarshalComposeFile(cf)
	output := string(yaml)

	// These should NOT appear in minimal YAML
	unwanted := []string{
		"deploy:",
		"healthcheck:",
		"logging:",
		"cap_add:",
		"cap_drop:",
		"privileged:",
		"read_only:",
		"dns:",
		"extra_hosts:",
		"hostname:",
		"network_mode:",
		"tmpfs:",
		"devices:",
		"sysctls:",
		"labels:",
		"environment:",
		"volumes:",
		"ports:",
		"networks:",
		"configs:",
		"secrets:",
	}

	for _, field := range unwanted {
		if strings.Contains(output, field) {
			t.Errorf("minimal YAML should not contain '%s':\n%s", field, output)
		}
	}
}

func TestBuildComposeFileYAMLValidity(t *testing.T) {
	// Build a full compose file and verify it produces valid YAML
	// that can be parsed back correctly
	d := makeResourceData(t, map[string]string{
		"name":                      "test-valid",
		"service.#":                 "1",
		"service.0.name":            "web",
		"service.0.image":           "nginx:latest",
		"service.0.ports.#":         "1",
		"service.0.ports.0":         "8080:80",
		"service.0.environment.%":   "1",
		"service.0.environment.FOO": "bar",
		"service.0.restart":         "always",
		"network.#":                 "1",
		"network.0.name":            "mynet",
		"network.0.driver":          "bridge",
		"volume.#":                  "1",
		"volume.0.name":             "myvol",
		"volume.0.driver":           "local",
	})

	cf := buildComposeFile(d)
	yamlBytes, err := docker.MarshalComposeFile(cf)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}

	// Parse it back
	cf2, err := docker.UnmarshalComposeFile(yamlBytes)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}

	// Verify parsed structure
	if cf2.Services["web"].Image != "nginx:latest" {
		t.Error("round-trip image mismatch")
	}
	if cf2.Services["web"].Restart != "always" {
		t.Error("round-trip restart mismatch")
	}
	if cf2.Services["web"].Ports[0] != "8080:80" {
		t.Error("round-trip ports mismatch")
	}
	if cf2.Services["web"].Environment["FOO"] != "bar" {
		t.Error("round-trip environment mismatch")
	}
	if cf2.Networks["mynet"].Driver != "bridge" {
		t.Error("round-trip network driver mismatch")
	}
	if cf2.Volumes["myvol"].Driver != "local" {
		t.Error("round-trip volume driver mismatch")
	}
}

func TestBuildComposeFileServiceWithProfiles(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                 "test-profiles",
		"service.#":            "1",
		"service.0.name":       "debug",
		"service.0.image":      "busybox:latest",
		"service.0.profiles.#": "2",
		"service.0.profiles.0": "debug",
		"service.0.profiles.1": "test",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["debug"]

	if len(svc.Profiles) != 2 || svc.Profiles[0] != "debug" || svc.Profiles[1] != "test" {
		t.Errorf("expected profiles [debug, test], got %v", svc.Profiles)
	}
}

func TestBuildComposeFileServiceWithEnvFile(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                 "test-envfile",
		"service.#":            "1",
		"service.0.name":       "app",
		"service.0.image":      "myapp:latest",
		"service.0.env_file.#": "2",
		"service.0.env_file.0": ".env",
		"service.0.env_file.1": ".env.production",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if len(svc.EnvFile) != 2 {
		t.Fatalf("expected 2 env_file entries, got %d", len(svc.EnvFile))
	}
	if svc.EnvFile[0] != ".env" {
		t.Errorf("expected .env, got %s", svc.EnvFile[0])
	}
}

func TestBuildComposeFileServiceWithNetworks(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                 "test-svc-nets",
		"service.#":            "1",
		"service.0.name":       "app",
		"service.0.image":      "myapp:latest",
		"service.0.networks.#": "2",
		"service.0.networks.0": "frontend",
		"service.0.networks.1": "backend",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if len(svc.Networks) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(svc.Networks))
	}
	if svc.Networks[0] != "frontend" || svc.Networks[1] != "backend" {
		t.Errorf("expected [frontend, backend], got %v", svc.Networks)
	}
}

func TestBuildComposeFileServiceWithExpose(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":               "test-expose",
		"service.#":          "1",
		"service.0.name":     "app",
		"service.0.image":    "myapp:latest",
		"service.0.expose.#": "2",
		"service.0.expose.0": "3000",
		"service.0.expose.1": "8080",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if len(svc.Expose) != 2 {
		t.Fatalf("expected 2 expose entries, got %d", len(svc.Expose))
	}
}

func TestBuildComposeFileServiceWithEntrypoint(t *testing.T) {
	d := makeResourceData(t, map[string]string{
		"name":                   "test-ep",
		"service.#":              "1",
		"service.0.name":         "app",
		"service.0.image":        "myapp:latest",
		"service.0.entrypoint.#": "2",
		"service.0.entrypoint.0": "/bin/sh",
		"service.0.entrypoint.1": "-c",
	})

	cf := buildComposeFile(d)
	svc := cf.Services["app"]

	if len(svc.Entrypoint) != 2 || svc.Entrypoint[0] != "/bin/sh" {
		t.Errorf("expected entrypoint [/bin/sh, -c], got %v", svc.Entrypoint)
	}
}
