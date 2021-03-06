package properties

	import "fmt"

type Bucket_InventoryConfiguration struct {
	
	
	
	
	
	
	
	Enabled interface{} `yaml:"Enabled"`
	Id interface{} `yaml:"Id"`
	IncludedObjectVersions interface{} `yaml:"IncludedObjectVersions"`
	Prefix interface{} `yaml:"Prefix,omitempty"`
	ScheduleFrequency interface{} `yaml:"ScheduleFrequency"`
	OptionalFields interface{} `yaml:"OptionalFields,omitempty"`
	Destination *Bucket_Destination `yaml:"Destination"`
}

func (resource Bucket_InventoryConfiguration) Validate() []error {
	errs := []error{}
	
	
	
	
	
	
	
	if resource.Enabled == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'Enabled'"))
	}
	if resource.Id == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'Id'"))
	}
	if resource.IncludedObjectVersions == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'IncludedObjectVersions'"))
	}
	if resource.ScheduleFrequency == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'ScheduleFrequency'"))
	}
	if resource.Destination == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'Destination'"))
	} else {
		errs = append(errs, resource.Destination.Validate()...)
	}
	return errs
}
