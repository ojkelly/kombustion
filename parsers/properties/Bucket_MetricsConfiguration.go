package properties

	import "fmt"

type Bucket_MetricsConfiguration struct {
	
	
	
	Id interface{} `yaml:"Id"`
	Prefix interface{} `yaml:"Prefix,omitempty"`
	TagFilters interface{} `yaml:"TagFilters,omitempty"`
}

func (resource Bucket_MetricsConfiguration) Validate() []error {
	errs := []error{}
	
	
	
	if resource.Id == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'Id'"))
	}
	return errs
}
