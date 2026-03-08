package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"terraform-provider-dockercompose/internal/docker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceComposeStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackCreate,
		Read:   resourceStackRead,
		Update: resourceStackUpdate,
		Delete: resourceStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project name for the Docker Compose stack. Used as the -p flag.",
			},
			"working_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Working directory for the stack. If not set, uses <project_directory>/<name>/.",
			},
			"remove_volumes_on_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to remove named volumes on destroy (docker compose down -v).",
			},

			// Computed outputs
			"compose_yaml": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The generated docker-compose YAML content.",
			},
			"compose_file_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Path to the generated docker-compose.yml file.",
			},

			// Container runtime info (populated after apply)
			"container": containerSchema(),

			// Service definitions
			"service": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Service definitions.",
				Elem: &schema.Resource{
					Schema: serviceSchema(),
				},
			},

			// Network definitions
			"network": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Network definitions.",
				Elem: &schema.Resource{
					Schema: networkSchema(),
				},
			},

			// Volume definitions
			"volume": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Volume definitions.",
				Elem: &schema.Resource{
					Schema: volumeSchema(),
				},
			},

			// Config definitions
			"config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Config definitions (Docker configs).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {Type: schema.TypeString, Required: true, Description: "Config name."},
						"file": {Type: schema.TypeString, Required: true, Description: "Path to the config file."},
					},
				},
			},

			// Secret definitions
			"secret": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Secret definitions (Docker secrets).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {Type: schema.TypeString, Required: true, Description: "Secret name."},
						"file": {Type: schema.TypeString, Required: true, Sensitive: true, Description: "Path to the secret file."},
					},
				},
			},
		},
	}
}

// ============================================================
// Schema definitions for nested blocks
// ============================================================

func serviceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// --- Core ---
		"name":           {Type: schema.TypeString, Required: true, Description: "Service name."},
		"image":          {Type: schema.TypeString, Required: true, Description: "Docker image."},
		"container_name": {Type: schema.TypeString, Optional: true, Description: "Custom container name."},
		"restart":        {Type: schema.TypeString, Optional: true, Description: "Restart policy: no, always, on-failure, unless-stopped."},
		"ports":          {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Port mappings (host:container)."},
		"expose":         {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Expose ports without publishing to host."},
		"depends_on":     {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Service dependencies."},
		"environment":    {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Environment variables."},
		"env_file":       {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Env file paths."},
		"command":        {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Override default command."},
		"entrypoint":     {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Override default entrypoint."},
		"volumes":        {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Volume mounts (host:container or named:container)."},
		"networks":       {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Networks to attach to."},
		"labels":         {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Container labels."},

		// --- Deploy ---
		"replicas":                     {Type: schema.TypeInt, Optional: true, Description: "Number of container replicas."},
		"resource_limits_cpus":         {Type: schema.TypeString, Optional: true, Description: "CPU limit (e.g. '0.5')."},
		"resource_limits_memory":       {Type: schema.TypeString, Optional: true, Description: "Memory limit (e.g. '512M')."},
		"resource_reservations_cpus":   {Type: schema.TypeString, Optional: true, Description: "CPU reservation."},
		"resource_reservations_memory": {Type: schema.TypeString, Optional: true, Description: "Memory reservation."},

		// --- Healthcheck ---
		"healthcheck_test":         {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Healthcheck command (e.g. ['CMD', 'curl', '-f', 'http://localhost'])."},
		"healthcheck_interval":     {Type: schema.TypeString, Optional: true, Description: "Healthcheck interval (e.g. '30s')."},
		"healthcheck_timeout":      {Type: schema.TypeString, Optional: true, Description: "Healthcheck timeout."},
		"healthcheck_retries":      {Type: schema.TypeInt, Optional: true, Description: "Healthcheck max retries."},
		"healthcheck_start_period": {Type: schema.TypeString, Optional: true, Description: "Healthcheck start period."},
		"healthcheck_disable":      {Type: schema.TypeBool, Optional: true, Default: false, Description: "Disable healthcheck."},

		// --- Logging ---
		"logging_driver":  {Type: schema.TypeString, Optional: true, Description: "Logging driver (json-file, syslog, etc.)."},
		"logging_options": {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Logging driver options."},

		// --- Security ---
		"cap_add":      {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Add Linux capabilities."},
		"cap_drop":     {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Drop Linux capabilities."},
		"security_opt": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Security options."},
		"privileged":   {Type: schema.TypeBool, Optional: true, Default: false, Description: "Run in privileged mode."},
		"read_only":    {Type: schema.TypeBool, Optional: true, Default: false, Description: "Mount root filesystem as read-only."},
		"init":         {Type: schema.TypeBool, Optional: true, Default: false, Description: "Run init process inside the container."},
		"user":         {Type: schema.TypeString, Optional: true, Description: "User to run as (user:group)."},

		// --- Networking ---
		"dns":          {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Custom DNS servers."},
		"extra_hosts":  {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Extra /etc/hosts entries (hostname:ip)."},
		"hostname":     {Type: schema.TypeString, Optional: true, Description: "Container hostname."},
		"domainname":   {Type: schema.TypeString, Optional: true, Description: "Container domain name."},
		"network_mode": {Type: schema.TypeString, Optional: true, Description: "Network mode (bridge, host, none, service:name)."},

		// --- Runtime ---
		"working_dir":       {Type: schema.TypeString, Optional: true, Description: "Working directory inside the container."},
		"stdin_open":        {Type: schema.TypeBool, Optional: true, Default: false, Description: "Keep stdin open (docker run -i)."},
		"tty":               {Type: schema.TypeBool, Optional: true, Default: false, Description: "Allocate pseudo-TTY (docker run -t)."},
		"shm_size":          {Type: schema.TypeString, Optional: true, Description: "Size of /dev/shm."},
		"stop_grace_period": {Type: schema.TypeString, Optional: true, Description: "Time to wait before force-stopping."},
		"stop_signal":       {Type: schema.TypeString, Optional: true, Description: "Signal to stop the container."},
		"platform":          {Type: schema.TypeString, Optional: true, Description: "Target platform (e.g. linux/amd64)."},
		"pull_policy":       {Type: schema.TypeString, Optional: true, Description: "Image pull policy (always, never, missing, build)."},
		"runtime":           {Type: schema.TypeString, Optional: true, Description: "Runtime (e.g. runc, nvidia)."},
		"tmpfs":             {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Tmpfs mounts."},
		"devices":           {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Device mappings."},
		"sysctls":           {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Sysctl settings."},
		"profiles":          {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Profiles this service belongs to."},
		"pid":               {Type: schema.TypeString, Optional: true, Description: "PID mode."},
		"ipc":               {Type: schema.TypeString, Optional: true, Description: "IPC mode."},
		"mem_limit":         {Type: schema.TypeString, Optional: true, Description: "Memory limit (legacy, prefer resource_limits_memory)."},
		"mem_reservation":   {Type: schema.TypeString, Optional: true, Description: "Memory reservation (legacy)."},
		"cpus":              {Type: schema.TypeString, Optional: true, Description: "CPU limit (legacy, prefer resource_limits_cpus)."},
	}
}

func networkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":         {Type: schema.TypeString, Required: true, Description: "Network name."},
		"driver":       {Type: schema.TypeString, Optional: true, Description: "Network driver (bridge, overlay, host, none)."},
		"driver_opts":  {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Driver-specific options."},
		"external":     {Type: schema.TypeBool, Optional: true, Default: false, Description: "Use externally created network."},
		"internal":     {Type: schema.TypeBool, Optional: true, Default: false, Description: "Restrict external access."},
		"attachable":   {Type: schema.TypeBool, Optional: true, Default: false, Description: "Allow manual container attachment."},
		"labels":       {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Network labels."},
		"ipam_driver":  {Type: schema.TypeString, Optional: true, Description: "IPAM driver."},
		"ipam_subnet":  {Type: schema.TypeString, Optional: true, Description: "IPAM subnet (e.g. '172.28.0.0/16')."},
		"ipam_gateway": {Type: schema.TypeString, Optional: true, Description: "IPAM gateway (e.g. '172.28.0.1')."},
	}
}

func volumeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":        {Type: schema.TypeString, Required: true, Description: "Volume name."},
		"driver":      {Type: schema.TypeString, Optional: true, Description: "Volume driver."},
		"driver_opts": {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Driver-specific options."},
		"external":    {Type: schema.TypeBool, Optional: true, Default: false, Description: "Use externally created volume."},
		"labels":      {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}, Description: "Volume labels."},
	}
}

// ============================================================
// CRUD Operations
// ============================================================

func resourceStackCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*docker.DockerClient)
	stackName := d.Get("name").(string)

	// Build the ComposeFile struct from Terraform config
	cf := buildComposeFile(d)

	// Marshal to YAML (deterministic key ordering via yaml.v2)
	yamlBytes, err := docker.MarshalComposeFile(cf)
	if err != nil {
		return fmt.Errorf("error generating compose YAML: %s", err)
	}

	// Determine project directory
	projectDir := client.ProjectDir(stackName)
	if wd, ok := d.GetOk("working_dir"); ok && wd.(string) != "" {
		projectDir = wd.(string)
	}

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("error creating project directory: %s", err)
	}

	composeFilePath := filepath.Join(projectDir, "docker-compose.yml")
	if err := os.WriteFile(composeFilePath, yamlBytes, 0644); err != nil {
		return fmt.Errorf("error writing compose file: %s", err)
	}

	// Run docker compose up
	if _, err := client.ComposeUp(stackName, composeFilePath); err != nil {
		return fmt.Errorf("error starting stack: %s", err)
	}

	d.SetId(stackName)
	if err := d.Set("compose_yaml", string(yamlBytes)); err != nil {
		return fmt.Errorf("error setting compose_yaml: %s", err)
	}
	if err := d.Set("compose_file_path", composeFilePath); err != nil {
		return fmt.Errorf("error setting compose_file_path: %s", err)
	}

	return resourceStackRead(d, m)
}

func resourceStackRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*docker.DockerClient)
	stackName := d.Id()

	composeFilePath := d.Get("compose_file_path").(string)
	if composeFilePath == "" {
		composeFilePath = client.ComposeFilePath(stackName)
	}

	// Restore compose file from state if missing on disk
	if _, err := os.Stat(composeFilePath); os.IsNotExist(err) {
		yamlContent := d.Get("compose_yaml").(string)
		if yamlContent != "" {
			if mkdirErr := os.MkdirAll(filepath.Dir(composeFilePath), 0755); mkdirErr != nil {
				return fmt.Errorf("error creating compose file directory: %s", mkdirErr)
			}
			if writeErr := os.WriteFile(composeFilePath, []byte(yamlContent), 0644); writeErr != nil {
				return fmt.Errorf("error restoring compose file from state: %s", writeErr)
			}
		} else {
			// No compose file on disk and nothing in state → resource gone
			d.SetId("")
			return nil
		}
	}

	// Verify services are running
	output, err := client.ComposePSServices(stackName, composeFilePath)
	if err != nil || strings.TrimSpace(output) == "" {
		// No running services → mark resource as destroyed
		d.SetId("")
		return nil
	}

	// Read container runtime info (IDs, IPs, ports, health, etc.)
	return readContainerInfo(d, client, stackName, composeFilePath)
}

func resourceStackUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceStackCreate(d, m)
}

func resourceStackDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*docker.DockerClient)
	stackName := d.Id()
	removeVolumes := d.Get("remove_volumes_on_destroy").(bool)

	composeFilePath := d.Get("compose_file_path").(string)
	if composeFilePath == "" {
		composeFilePath = client.ComposeFilePath(stackName)
	}

	if _, err := client.ComposeDown(stackName, composeFilePath, removeVolumes); err != nil {
		return fmt.Errorf("error stopping stack: %s", err)
	}

	// Clean up generated files
	os.Remove(composeFilePath)
	os.Remove(filepath.Dir(composeFilePath)) // Only succeeds if dir is empty

	return nil
}

// ============================================================
// Compose File Builder
// ============================================================

func buildComposeFile(d *schema.ResourceData) *docker.ComposeFile {
	cf := &docker.ComposeFile{
		Services: make(map[string]*docker.ServiceConfig),
	}

	// --- Services ---
	rawServices := d.Get("service").([]interface{})
	for _, raw := range rawServices {
		svc := raw.(map[string]interface{})
		name := svc["name"].(string)

		config := &docker.ServiceConfig{
			Image:           svc["image"].(string),
			ContainerName:   getStr(svc, "container_name"),
			Restart:         getStr(svc, "restart"),
			Ports:           getStrList(svc, "ports"),
			Expose:          getStrList(svc, "expose"),
			DependsOn:       getStrList(svc, "depends_on"),
			Environment:     getStrMap(svc, "environment"),
			EnvFile:         getStrList(svc, "env_file"),
			Command:         getStrList(svc, "command"),
			Entrypoint:      getStrList(svc, "entrypoint"),
			Volumes:         getStrList(svc, "volumes"),
			Networks:        getStrList(svc, "networks"),
			Labels:          getStrMap(svc, "labels"),
			CapAdd:          getStrList(svc, "cap_add"),
			CapDrop:         getStrList(svc, "cap_drop"),
			SecurityOpt:     getStrList(svc, "security_opt"),
			Privileged:      getBool(svc, "privileged"),
			ReadOnly:        getBool(svc, "read_only"),
			Init:            getBoolPtr(svc, "init"),
			User:            getStr(svc, "user"),
			DNS:             getStrList(svc, "dns"),
			ExtraHosts:      getStrList(svc, "extra_hosts"),
			Hostname:        getStr(svc, "hostname"),
			Domainname:      getStr(svc, "domainname"),
			NetworkMode:     getStr(svc, "network_mode"),
			WorkingDir:      getStr(svc, "working_dir"),
			StdinOpen:       getBool(svc, "stdin_open"),
			Tty:             getBool(svc, "tty"),
			ShmSize:         getStr(svc, "shm_size"),
			StopGracePeriod: getStr(svc, "stop_grace_period"),
			StopSignal:      getStr(svc, "stop_signal"),
			Platform:        getStr(svc, "platform"),
			PullPolicy:      getStr(svc, "pull_policy"),
			Runtime:         getStr(svc, "runtime"),
			Tmpfs:           getStrList(svc, "tmpfs"),
			Devices:         getStrList(svc, "devices"),
			Sysctls:         getStrMap(svc, "sysctls"),
			Profiles:        getStrList(svc, "profiles"),
			Pid:             getStr(svc, "pid"),
			Ipc:             getStr(svc, "ipc"),
			MemLimit:        getStr(svc, "mem_limit"),
			MemReservation:  getStr(svc, "mem_reservation"),
			Cpus:            getStr(svc, "cpus"),
		}

		// Build deploy config if any deploy-related field is set
		replicas := getIntPtr(svc, "replicas")
		limCpus := getStr(svc, "resource_limits_cpus")
		limMem := getStr(svc, "resource_limits_memory")
		resCpus := getStr(svc, "resource_reservations_cpus")
		resMem := getStr(svc, "resource_reservations_memory")

		hasLimits := limCpus != "" || limMem != ""
		hasReservations := resCpus != "" || resMem != ""
		hasDeploy := replicas != nil || hasLimits || hasReservations

		if hasDeploy {
			config.Deploy = &docker.DeployConfig{Replicas: replicas}

			if hasLimits || hasReservations {
				config.Deploy.Resources = &docker.DeployResources{}
				if hasLimits {
					config.Deploy.Resources.Limits = &docker.ResourceSpec{Cpus: limCpus, Memory: limMem}
				}
				if hasReservations {
					config.Deploy.Resources.Reservations = &docker.ResourceSpec{Cpus: resCpus, Memory: resMem}
				}
			}
		}

		// Build healthcheck if any healthcheck field is set
		hcTest := getStrList(svc, "healthcheck_test")
		hcInterval := getStr(svc, "healthcheck_interval")
		hcTimeout := getStr(svc, "healthcheck_timeout")
		hcRetries := getIntPtr(svc, "healthcheck_retries")
		hcStart := getStr(svc, "healthcheck_start_period")
		hcDisable := getBool(svc, "healthcheck_disable")

		if len(hcTest) > 0 || hcInterval != "" || hcTimeout != "" || hcRetries != nil || hcStart != "" || hcDisable {
			config.Healthcheck = &docker.HealthcheckCfg{
				Test:        hcTest,
				Interval:    hcInterval,
				Timeout:     hcTimeout,
				Retries:     hcRetries,
				StartPeriod: hcStart,
				Disable:     hcDisable,
			}
		}

		// Build logging config if logging fields are set
		logDriver := getStr(svc, "logging_driver")
		logOpts := getStrMap(svc, "logging_options")
		if logDriver != "" || len(logOpts) > 0 {
			config.Logging = &docker.LoggingConfig{
				Driver:  logDriver,
				Options: logOpts,
			}
		}

		cf.Services[name] = config
	}

	// --- Networks ---
	if rawNets, ok := d.GetOk("network"); ok {
		cf.Networks = make(map[string]*docker.NetworkConfig)
		for _, raw := range rawNets.([]interface{}) {
			net := raw.(map[string]interface{})
			name := net["name"].(string)

			nc := &docker.NetworkConfig{
				Driver:     getStr(net, "driver"),
				DriverOpts: getStrMap(net, "driver_opts"),
				External:   getBool(net, "external"),
				Internal:   getBool(net, "internal"),
				Attachable: getBool(net, "attachable"),
				Labels:     getStrMap(net, "labels"),
			}

			ipamDriver := getStr(net, "ipam_driver")
			ipamSubnet := getStr(net, "ipam_subnet")
			ipamGateway := getStr(net, "ipam_gateway")
			if ipamDriver != "" || ipamSubnet != "" || ipamGateway != "" {
				nc.IPAM = &docker.IPAMConfig{Driver: ipamDriver}
				if ipamSubnet != "" || ipamGateway != "" {
					nc.IPAM.Config = []docker.IPAMPool{{
						Subnet:  ipamSubnet,
						Gateway: ipamGateway,
					}}
				}
			}

			cf.Networks[name] = nc
		}
	}

	// --- Volumes ---
	if rawVols, ok := d.GetOk("volume"); ok {
		cf.Volumes = make(map[string]*docker.VolumeConfig)
		for _, raw := range rawVols.([]interface{}) {
			vol := raw.(map[string]interface{})
			name := vol["name"].(string)
			cf.Volumes[name] = &docker.VolumeConfig{
				Driver:     getStr(vol, "driver"),
				DriverOpts: getStrMap(vol, "driver_opts"),
				External:   getBool(vol, "external"),
				Labels:     getStrMap(vol, "labels"),
			}
		}
	}

	// --- Configs ---
	if rawConfigs, ok := d.GetOk("config"); ok {
		cf.Configs = make(map[string]*docker.ConfigEntry)
		for _, raw := range rawConfigs.([]interface{}) {
			cfg := raw.(map[string]interface{})
			cf.Configs[cfg["name"].(string)] = &docker.ConfigEntry{
				File: cfg["file"].(string),
			}
		}
	}

	// --- Secrets ---
	if rawSecrets, ok := d.GetOk("secret"); ok {
		cf.Secrets = make(map[string]*docker.SecretEntry)
		for _, raw := range rawSecrets.([]interface{}) {
			sec := raw.(map[string]interface{})
			cf.Secrets[sec["name"].(string)] = &docker.SecretEntry{
				File: sec["file"].(string),
			}
		}
	}

	return cf
}
