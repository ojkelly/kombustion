package properties

	import "fmt"

type AutoScalingGroup_NotificationConfiguration struct {
	
	
	TopicARN interface{} `yaml:"TopicARN"`
	NotificationTypes interface{} `yaml:"NotificationTypes,omitempty"`
}

func (resource AutoScalingGroup_NotificationConfiguration) Validate() []error {
	errs := []error{}
	
	
	if resource.TopicARN == nil {
		errs = append(errs, fmt.Errorf("Missing required field 'TopicARN'"))
	}
	return errs
}