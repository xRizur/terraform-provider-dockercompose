package docker

import (
	"gopkg.in/yaml.v2"
)

// ComposeFile represents a complete docker-compose.yml structure.
type ComposeFile struct {
	Services map[string]*ServiceConfig `yaml:"services"`
	Networks map[string]*NetworkConfig `yaml:"networks,omitempty"`
	Volumes  map[string]*VolumeConfig  `yaml:"volumes,omitempty"`
	Configs  map[string]*ConfigEntry   `yaml:"configs,omitempty"`
	Secrets  map[string]*SecretEntry   `yaml:"secrets,omitempty"`
}

// ServiceConfig represents a single service in docker-compose.
type ServiceConfig struct {
	Image           string            `yaml:"image,omitempty"`
	ContainerName   string            `yaml:"container_name,omitempty"`
	Restart         string            `yaml:"restart,omitempty"`
	Ports           []string          `yaml:"ports,omitempty"`
	Expose          []string          `yaml:"expose,omitempty"`
	DependsOn       []string          `yaml:"depends_on,omitempty"`
	Environment     map[string]string `yaml:"environment,omitempty"`
	EnvFile         []string          `yaml:"env_file,omitempty"`
	Command         []string          `yaml:"command,omitempty"`
	Entrypoint      []string          `yaml:"entrypoint,omitempty"`
	Volumes         []string          `yaml:"volumes,omitempty"`
	Networks        []string          `yaml:"networks,omitempty"`
	Labels          map[string]string `yaml:"labels,omitempty"`
	Deploy          *DeployConfig     `yaml:"deploy,omitempty"`
	Healthcheck     *HealthcheckCfg   `yaml:"healthcheck,omitempty"`
	Logging         *LoggingConfig    `yaml:"logging,omitempty"`
	CapAdd          []string          `yaml:"cap_add,omitempty"`
	CapDrop         []string          `yaml:"cap_drop,omitempty"`
	SecurityOpt     []string          `yaml:"security_opt,omitempty"`
	Privileged      bool              `yaml:"privileged,omitempty"`
	ReadOnly        bool              `yaml:"read_only,omitempty"`
	Init            *bool             `yaml:"init,omitempty"`
	User            string            `yaml:"user,omitempty"`
	DNS             []string          `yaml:"dns,omitempty"`
	ExtraHosts      []string          `yaml:"extra_hosts,omitempty"`
	Hostname        string            `yaml:"hostname,omitempty"`
	Domainname      string            `yaml:"domainname,omitempty"`
	NetworkMode     string            `yaml:"network_mode,omitempty"`
	WorkingDir      string            `yaml:"working_dir,omitempty"`
	StdinOpen       bool              `yaml:"stdin_open,omitempty"`
	Tty             bool              `yaml:"tty,omitempty"`
	ShmSize         string            `yaml:"shm_size,omitempty"`
	StopGracePeriod string            `yaml:"stop_grace_period,omitempty"`
	StopSignal      string            `yaml:"stop_signal,omitempty"`
	Platform        string            `yaml:"platform,omitempty"`
	PullPolicy      string            `yaml:"pull_policy,omitempty"`
	Runtime         string            `yaml:"runtime,omitempty"`
	Tmpfs           []string          `yaml:"tmpfs,omitempty"`
	Devices         []string          `yaml:"devices,omitempty"`
	Sysctls         map[string]string `yaml:"sysctls,omitempty"`
	Profiles        []string          `yaml:"profiles,omitempty"`
	Pid             string            `yaml:"pid,omitempty"`
	Ipc             string            `yaml:"ipc,omitempty"`
	MemLimit        string            `yaml:"mem_limit,omitempty"`
	MemReservation  string            `yaml:"mem_reservation,omitempty"`
	Cpus            string            `yaml:"cpus,omitempty"`
}

// DeployConfig represents the deploy section of a service.
type DeployConfig struct {
	Replicas  *int             `yaml:"replicas,omitempty"`
	Resources *DeployResources `yaml:"resources,omitempty"`
}

// DeployResources represents resource constraints.
type DeployResources struct {
	Limits       *ResourceSpec `yaml:"limits,omitempty"`
	Reservations *ResourceSpec `yaml:"reservations,omitempty"`
}

// ResourceSpec represents CPU/memory constraints.
type ResourceSpec struct {
	Cpus   string `yaml:"cpus,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

// HealthcheckCfg represents a service healthcheck.
type HealthcheckCfg struct {
	Test        []string `yaml:"test,omitempty"`
	Interval    string   `yaml:"interval,omitempty"`
	Timeout     string   `yaml:"timeout,omitempty"`
	Retries     *int     `yaml:"retries,omitempty"`
	StartPeriod string   `yaml:"start_period,omitempty"`
	Disable     bool     `yaml:"disable,omitempty"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Driver  string            `yaml:"driver,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

// NetworkConfig represents a top-level network definition.
type NetworkConfig struct {
	Name       string            `yaml:"name,omitempty"`
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty"`
	Internal   bool              `yaml:"internal,omitempty"`
	Attachable bool              `yaml:"attachable,omitempty"`
	IPAM       *IPAMConfig       `yaml:"ipam,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty"`
}

// IPAMConfig represents IPAM configuration for a network.
type IPAMConfig struct {
	Driver string     `yaml:"driver,omitempty"`
	Config []IPAMPool `yaml:"config,omitempty"`
}

// IPAMPool represents a single IPAM address pool.
type IPAMPool struct {
	Subnet  string `yaml:"subnet,omitempty"`
	Gateway string `yaml:"gateway,omitempty"`
}

// VolumeConfig represents a top-level volume definition.
type VolumeConfig struct {
	Name       string            `yaml:"name,omitempty"`
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty"`
}

// ConfigEntry represents a top-level config definition.
type ConfigEntry struct {
	File string `yaml:"file,omitempty"`
}

// SecretEntry represents a top-level secret definition.
type SecretEntry struct {
	File string `yaml:"file,omitempty"`
}

// MarshalComposeFile serializes a ComposeFile to YAML bytes.
func MarshalComposeFile(cf *ComposeFile) ([]byte, error) {
	return yaml.Marshal(cf)
}

// UnmarshalComposeFile deserializes YAML bytes into a ComposeFile.
func UnmarshalComposeFile(data []byte) (*ComposeFile, error) {
	cf := &ComposeFile{}
	err := yaml.Unmarshal(data, cf)
	return cf, err
}
