package properties


type TaskDefinition_ContainerDefinition struct {
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	Cpu interface{} `yaml:"Cpu,omitempty"`
	DisableNetworking interface{} `yaml:"DisableNetworking,omitempty"`
	Essential interface{} `yaml:"Essential,omitempty"`
	Hostname interface{} `yaml:"Hostname,omitempty"`
	Image interface{} `yaml:"Image,omitempty"`
	Memory interface{} `yaml:"Memory,omitempty"`
	MemoryReservation interface{} `yaml:"MemoryReservation,omitempty"`
	Name interface{} `yaml:"Name,omitempty"`
	Privileged interface{} `yaml:"Privileged,omitempty"`
	ReadonlyRootFilesystem interface{} `yaml:"ReadonlyRootFilesystem,omitempty"`
	User interface{} `yaml:"User,omitempty"`
	WorkingDirectory interface{} `yaml:"WorkingDirectory,omitempty"`
	DockerLabels interface{} `yaml:"DockerLabels,omitempty"`
	LogConfiguration *TaskDefinition_LogConfiguration `yaml:"LogConfiguration,omitempty"`
	Command interface{} `yaml:"Command,omitempty"`
	DnsServers interface{} `yaml:"DnsServers,omitempty"`
	DockerSecurityOptions interface{} `yaml:"DockerSecurityOptions,omitempty"`
	EntryPoint interface{} `yaml:"EntryPoint,omitempty"`
	Environment interface{} `yaml:"Environment,omitempty"`
	ExtraHosts interface{} `yaml:"ExtraHosts,omitempty"`
	Links interface{} `yaml:"Links,omitempty"`
	Ulimits interface{} `yaml:"Ulimits,omitempty"`
	DnsSearchDomains interface{} `yaml:"DnsSearchDomains,omitempty"`
	MountPoints interface{} `yaml:"MountPoints,omitempty"`
	PortMappings interface{} `yaml:"PortMappings,omitempty"`
	VolumesFrom interface{} `yaml:"VolumesFrom,omitempty"`
	LinuxParameters *TaskDefinition_LinuxParameters `yaml:"LinuxParameters,omitempty"`
}

func (resource TaskDefinition_ContainerDefinition) Validate() []error {
	errs := []error{}
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	return errs
}
