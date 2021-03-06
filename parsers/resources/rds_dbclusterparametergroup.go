package resources

import (
	yaml "github.com/KablamoOSS/yaml"
	"github.com/KablamoOSS/kombustion/types"
	"log"
	"fmt"
)

type RDSDBClusterParameterGroup struct {
	Type       string                      `yaml:"Type"`
	Properties RDSDBClusterParameterGroupProperties `yaml:"Properties"`
	Condition  interface{}                 `yaml:"Condition,omitempty"`
	Metadata   interface{}                 `yaml:"Metadata,omitempty"`
	DependsOn  interface{}                 `yaml:"DependsOn,omitempty"`
}

type RDSDBClusterParameterGroupProperties struct {
	Description interface{} `yaml:"Description"`
	Family interface{} `yaml:"Family"`
	Parameters interface{} `yaml:"Parameters"`
	Tags interface{} `yaml:"Tags,omitempty"`
}

func NewRDSDBClusterParameterGroup(properties RDSDBClusterParameterGroupProperties, deps ...interface{}) RDSDBClusterParameterGroup {
	return RDSDBClusterParameterGroup{
		Type:       "AWS::RDS::DBClusterParameterGroup",
		Properties: properties,
		DependsOn:  deps,
	}
}

func ParseRDSDBClusterParameterGroup(name string, data string) (cf types.ValueMap, err error) {
	var resource RDSDBClusterParameterGroup
	if err = yaml.Unmarshal([]byte(data), &resource); err != nil {
		return
	}
	if errs := resource.Properties.Validate(); len(errs) > 0 {
		for _, err = range errs {
			log.Println("WARNING: RDSDBClusterParameterGroup - ", err)
		}
		return
	}
	cf = types.ValueMap{name: resource}
	return
}

func (resource RDSDBClusterParameterGroup) Validate() []error {
	return resource.Properties.Validate()
}

func (resource RDSDBClusterParameterGroupProperties) Validate() []error {
	errs := []error{}
	if resource.Description == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'Description'"))
	}
	if resource.Family == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'Family'"))
	}
	if resource.Parameters == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'Parameters'"))
	}
	return errs
}
