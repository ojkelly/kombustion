package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

/*
	generate
	Auto generates go parser code from the cloudformation spec
*/

type CfnSpec struct {
	PropertyTypes map[string]CfnType
	ResourceTypes map[string]CfnType
}

type CfnType struct {
	Documentation string
	Properties    map[string]CfnProperty
	Attributes    map[string]CfnAttribute
}

type CfnAttribute struct {
	PrimitiveType string
}

type CfnProperty struct {
	Documentation     string
	Type              string
	PrimitiveType     string
	ItemType          string
	Required          bool
	DuplicatesAllowed bool
	UpdateType        string
}

type NamedCfnProperty struct {
	CfnProperty
	name string
}

const sourceDir = "./generate/source/"

var mainPackageName string
var parsersDir string
var propertiesDir string
var outputsDir string
var resourcesDir string

var cfnEndpoints = map[string]string{
	"Sydney":           "https://d2stg8d246z9di.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Singapore":        "https://doigdx0kgq9el.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Mumbai":           "https://d2senuesg1djtx.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Seoul":            "https://d1ane3fvebulky.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Tokyo":            "https://d33vqc0rt9ld30.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Canada":           "https://d2s8ygphhesbe7.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Frankfurt":        "https://d1mta8qj7i28i2.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"London":           "https://d1742qcu2c1ncx.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Ireland":          "https://d3teyb21fexa9r.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Sao Paulo":        "https://d3c9jyj3w509b0.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"North Virginia":   "https://d1uauaxba7bl26.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Ohio":             "https://dnwj8swjjbsbt.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"North California": "https://d68hl49wbnanq.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
	"Oregon":           "https://d201a2mn26r7lk.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json",
}

const parserMapTemplate = `{{$MainPackageName := .MainPackageName}}package {{$MainPackageName}}
{{$PackageName := .PackageName}}

import (
  "github.com/KablamoOSS/kombustion/types"
  "github.com/KablamoOSS/kombustion/{{$MainPackageName}}/{{$PackageName}}"
)

func GetParsers_{{$PackageName}}() map[string]types.ParserFunc {
	return map[string]types.ParserFunc{
		{{range $ResourceType, $ResourceName := .ResourceTypes}}
		"{{$ResourceType}}": {{$PackageName}}.Parse{{$ResourceName}},
		{{end}}
	}
}
`

const propertyTemplate = `package properties
{{$PropertyName := .PropertyName}}

{{- if .NeedsFmtImport}}
	import "fmt"
{{- end}}

type {{$PropertyName}} struct {
	{{- range $property := .PropertyStrings}}
	{{$property}}
	{{- end}}
}

func (resource {{$PropertyName}}) Validate() []error {
	errs := []error{}
	{{- range $validator := .ValidatorStrings}}
	{{$validator}}
	{{- end}}
	return errs
}
`

const resourceTemplate = `package resources
{{$MainPackageName := .MainPackageName}}
{{$ResourceName := .ResourceName -}}
{{- $BT := "` + "`" + `" -}}

import (
	yaml "github.com/KablamoOSS/yaml"
	"github.com/KablamoOSS/kombustion/types"
	"log"
	{{- if .NeedsFmtImport}}
	"fmt"
	{{- end}}
	{{- if .NeedsPropertiesImport}}
	"github.com/KablamoOSS/kombustion/{{$MainPackageName}}/properties"
	{{- end}}
)

type {{$ResourceName}} struct {
	Type       string                      {{$BT}}yaml:"Type"{{$BT}}
	Properties {{$ResourceName}}Properties {{$BT}}yaml:"Properties"{{$BT}}
	Condition  interface{}                 {{$BT}}yaml:"Condition,omitempty"{{$BT}}
	Metadata   interface{}                 {{$BT}}yaml:"Metadata,omitempty"{{$BT}}
	DependsOn  interface{}                 {{$BT}}yaml:"DependsOn,omitempty"{{$BT}}
}

type {{$ResourceName}}Properties struct {
	{{- range $property := .PropertyStrings}}
	{{$property}}
	{{- end}}
}

func New{{$ResourceName}}(properties {{$ResourceName}}Properties, deps ...interface{}) {{$ResourceName}} {
	return {{$ResourceName}}{
		Type:       "{{.Type}}",
		Properties: properties,
		DependsOn:  deps,
	}
}

func Parse{{$ResourceName}}(name string, data string) (cf types.ValueMap, err error) {
	var resource {{$ResourceName}}
	if err = yaml.Unmarshal([]byte(data), &resource); err != nil {
		return
	}
	if errs := resource.Properties.Validate(); len(errs) > 0 {
		for _, err = range errs {
			log.Println("WARNING: {{$ResourceName}} - ", err)
		}
		return
	}
	cf = types.ValueMap{name: resource}
	return
}

func (resource {{$ResourceName}}) Validate() []error {
	return resource.Properties.Validate()
}

func (resource {{$ResourceName}}Properties) Validate() []error {
	errs := []error{}
	{{- range $validator := .ValidatorStrings}}
	{{$validator}}
	{{- end}}
	return errs
}
`

const outputTemplate = `package outputs
{{$ResourceName := .ResourceName -}}

import (
	{{- if .Attributes}}
	yaml "github.com/KablamoOSS/yaml"
	{{- end}}
	"github.com/KablamoOSS/kombustion/types"
)

func Parse{{$ResourceName}}(name string, data string) (cf types.ValueMap, err error) {
	{{if .Attributes}}
	var resource, output types.ValueMap
	if err = yaml.Unmarshal([]byte(data), &resource); err != nil {
		return
	}
	{{end}}
	cf = types.ValueMap{
		name: types.ValueMap{
			"Description": name + " Object",
			"Value": map[string]interface{}{
				"Ref": name,
			},
			"Export": map[string]interface{}{
				"Name": map[string]interface{}{
					"Fn::Sub": "${AWS::StackName}-{{$ResourceName}}-" + name,
				},
			},
		},
	}

	{{range $attrName, $attr := .Attributes}}
	output = types.ValueMap{
		"Description": name + " Object",
		"Value": map[string]interface{}{
			"Fn::GetAtt": []string{name, "{{$attrName}}"},
		},
		"Export": map[string]interface{}{
			"Name": map[string]interface{}{
				"Fn::Sub": "${AWS::StackName}-{{$ResourceName}}-" + name + "-{{$attr}}",
			},
		},
	}
	if condition, ok := resource["Condition"]; ok {
		output["Condition"] = condition
	}
	cf[name+"{{$attr}}"] = output
	{{end}}

	return
}
`

const validatorTemplate = `
	{{- if .PrimitiveType -}}
	if resource.{{.Name}} == nil {
		errs = append(errs, fmt.Errorf("Missing required field '{{.Name}}'"))
	}
	{{- else if .ListMapType -}}
	if resource.{{.Name}} == nil {
		errs = append(errs, fmt.Errorf("Missing required field '{{.Name}}'"))
	} else {
		errs = append(errs, resource.{{.Name}}.Validate()...)
	}
	{{- else -}}
	if resource.{{.Name}} == nil {
		errs = append(errs, fmt.Errorf("Missing required field '{{.Name}}'"))
	}
	{{- end -}}
`

var globalPropertyTypes []string

func init() {
	mainPackageName = "parsers"
	if len(os.Args) > 1 {
		mainPackageName = os.Args[1]
	}

	parsersDir = fmt.Sprintf("./%v/", mainPackageName)
	propertiesDir = fmt.Sprintf("./%v/properties/", mainPackageName)
	outputsDir = fmt.Sprintf("./%v/outputs/", mainPackageName)
	resourcesDir = fmt.Sprintf("./%v/resources/", mainPackageName)

	os.Mkdir(sourceDir, 0744)
	os.Mkdir(parsersDir, 0744)
	os.Mkdir(outputsDir, 0744)
	os.Mkdir(resourcesDir, 0744)
	os.Mkdir(propertiesDir, 0744)
}

func main() {

	var cfnSpec CfnSpec

	checkLocalData(sourceDir, cfnEndpoints)
	cfnSpec = buildUniqueSet(sourceDir, cfnEndpoints)

	buildJsonParsers(cfnSpec)
	buildYamlParsers(cfnSpec)

	// filePath := fmt.Sprintf("%vtypes.go", typesDir)
	// err := ioutil.WriteFile(filePath, []byte(typesData), 0644)
	//checkError(err)
}

func checkLocalData(sourceDir string, specList map[string]string) {
	for _, region := range sortSpecList(specList) {
		url := specList[region]
		if _, err := os.Stat(sourceDir + "/" + region + ".json"); os.IsNotExist(err) {
			cfnData := fetchSourceData(region)
			if len(cfnData) > 0 {
				log.Println("Downloaded region: " + region + " from " + url)
			}
		}
	}
}

func buildUniqueSet(sourceDir string, specList map[string]string) CfnSpec {
	uniquecfnData := CfnSpec{
		PropertyTypes: map[string]CfnType{},
		ResourceTypes: map[string]CfnType{},
	}
	for _, region := range sortSpecList(specList) {
		var tempcfnSpec CfnSpec
		cfnData, err := ioutil.ReadFile(fmt.Sprintf("%v%v.json", sourceDir, region))
		if err == nil {
			err = json.Unmarshal(cfnData, &tempcfnSpec)
			if err == nil {
				for k, v := range tempcfnSpec.PropertyTypes {
					if _, ok := uniquecfnData.PropertyTypes[k]; ok {
					} else {
						uniquecfnData.PropertyTypes[k] = v
					}
				}
				for k, v := range tempcfnSpec.ResourceTypes {
					if _, ok := uniquecfnData.ResourceTypes[k]; ok {
					} else {
						uniquecfnData.ResourceTypes[k] = v
					}
				}
			}
		}
	}
	return uniquecfnData
}

func fetchSourceData(region string) []byte {
	cfnUrl := cfnEndpoints[region]
	request, err := http.NewRequest("GET", cfnUrl, nil)
	checkError(err)

	request.Header.Set("Content-Type", "application/json")
	response, err := new(http.Client).Do(request)
	checkError(err)

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch cfn template data: %v", response.StatusCode)
	}

	cfnData, err := ioutil.ReadAll(response.Body)
	checkError(err)

	err = ioutil.WriteFile(fmt.Sprintf("%v%v.json", sourceDir, region), cfnData, 0644)
	checkError(err)

	return cfnData
}

func buildJsonParsers(cfnSpec CfnSpec) {
	//TODO
}

func buildYamlParsers(cfnSpec CfnSpec) {
	// check for global types
	globalPropertyTypes := []string{}
	for k := range cfnSpec.PropertyTypes {
		if isPropertyGlobal(k) {
			globalPropertyTypes = append(globalPropertyTypes, k)
		}
	}

	resourceParsersObject := buildParserMapping(cfnSpec, "resources")
	filePath := fmt.Sprintf("%vresources.go", parsersDir)
	err := ioutil.WriteFile(filePath, []byte(resourceParsersObject), 0644)
	checkError(err)

	outputParsersObject := buildParserMapping(cfnSpec, "outputs")
	filePath = fmt.Sprintf("%voutput.go", parsersDir)
	err = ioutil.WriteFile(filePath, []byte(outputParsersObject), 0644)
	checkError(err)

	// properties
	for k, cfnType := range cfnSpec.PropertyTypes {
		propertyObject := buildPropertyYaml(k, cfnType)
		filePath := fmt.Sprintf("%v%v.go", propertiesDir, propertyNameFromPropertyType(k))
		err := ioutil.WriteFile(filePath, []byte(propertyObject), 0644)
		checkError(err)
	}

	// Resource parsers
	for k, cfnType := range cfnSpec.ResourceTypes {
		resourceObject := buildResourceYaml(k, cfnType)
		filePath := fmt.Sprintf("%v%v.go", resourcesDir, fileNameFromCfnType(k))
		err := ioutil.WriteFile(filePath, []byte(resourceObject), 0644)
		checkError(err)
	}

	// Output parsers
	for k, cfnType := range cfnSpec.ResourceTypes {
		outputObject := buildOutputYaml(k, cfnType)
		filePath := fmt.Sprintf("%v%v.go", outputsDir, fileNameFromCfnType(k))
		err := ioutil.WriteFile(filePath, []byte(outputObject), 0644)
		checkError(err)
	}
}

func buildParserMapping(cfnSpec CfnSpec, packageName string) string {
	resourceTypes := make(map[string]string)
	for _, k := range sortTypeNames(cfnSpec.ResourceTypes) {
		name := titleCaseNameFromCfnType(k)
		resourceTypes[k] = name
	}

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("").Parse(parserMapTemplate))
	err := t.Execute(buf, map[string]interface{}{
		"PackageName":     packageName,
		"ResourceTypes":   resourceTypes,
		"MainPackageName": mainPackageName,
	})
	checkError(err)
	return buf.String()
}

func buildPropertyYaml(obj string, cfnType CfnType) string {
	propertyStrings := make([]string, len(cfnType.Properties))
	validatorStrings := make([]string, len(cfnType.Properties))
	for _, property := range sortProperties(cfnType.Properties) {
		if str := valueStringYaml("", obj, property.name, property.CfnProperty); len(str) > 0 {
			propertyStrings = append(propertyStrings, str)
		}
	}
	for _, property := range sortProperties(cfnType.Properties) {
		if str := validatorYaml(obj, property.name, property.CfnProperty); len(str) > 0 {
			validatorStrings = append(validatorStrings, str)
		}
	}

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("").Parse(propertyTemplate))
	err := t.Execute(buf, map[string]interface{}{
		"PropertyName":     propertyNameFromPropertyType(obj),
		"PropertyStrings":  propertyStrings,
		"ValidatorStrings": validatorStrings,
		"NeedsFmtImport":   needsFmtImport(cfnType),
	})
	checkError(err)
	return buf.String()
}

func buildResourceYaml(obj string, cfnType CfnType) string {
	propertyStrings := make([]string, 0, len(cfnType.Properties))
	validatorStrings := make([]string, 0, len(cfnType.Properties))
	for _, property := range sortProperties(cfnType.Properties) {
		if str := valueStringYaml("properties.", obj, property.name, property.CfnProperty); len(str) > 0 {
			propertyStrings = append(propertyStrings, str)
		}
	}
	for _, property := range sortProperties(cfnType.Properties) {
		if str := validatorYaml(obj, property.name, property.CfnProperty); len(str) > 0 {
			validatorStrings = append(validatorStrings, str)
		}
	}

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("").Parse(resourceTemplate))
	err := t.Execute(buf, map[string]interface{}{
		"Type":                  obj,
		"ResourceName":          titleCaseNameFromCfnType(obj),
		"PropertyStrings":       propertyStrings,
		"ValidatorStrings":      validatorStrings,
		"NeedsFmtImport":        needsFmtImport(cfnType),
		"NeedsPropertiesImport": needsPropertiesImport(cfnType),
		"MainPackageName":       mainPackageName,
	})
	checkError(err)
	return buf.String()
}

func buildOutputYaml(obj string, cfnType CfnType) string {
	alnumRegex, _ := regexp.Compile("[^a-zA-Z0-9]+")
	attributes := make(map[string]string)
	for _, attName := range sortAttributeNames(cfnType.Attributes) {
		attributes[attName] = alnumRegex.ReplaceAllString(attName, "")
	}

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("").Parse(outputTemplate))
	err := t.Execute(buf, map[string]interface{}{
		"ResourceName": titleCaseNameFromCfnType(obj),
		"Attributes":   attributes,
	})
	checkError(err)
	return buf.String()
}

func valueStringYaml(propPackage, obj, name string, property CfnProperty) string {
	omitempty := ",omitempty"
	if property.Required {
		omitempty = ""
	}
	if len(property.PrimitiveType) > 0 {
		return name + " interface{} `yaml:" + `"` + name + omitempty + `"` + "`"
	} else if len(property.Type) > 0 && property.Type != "List" && property.Type != "Map" {
		subPropertyName := propertyNameFromResourceType(obj, property.Type)
		return name + " *" + propPackage + subPropertyName + " `yaml:" + `"` + name + omitempty + `"` + "`"
	}
	return name + " interface{} `yaml:" + `"` + name + omitempty + `"` + "`"
}

func validatorYaml(obj, name string, property CfnProperty) string {
	if !property.Required {
		return ""
	}

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("").Parse(validatorTemplate))
	err := t.Execute(buf, map[string]interface{}{
		"Name":          name,
		"PrimitiveType": len(property.PrimitiveType) > 0,
		"ListMapType":   len(property.Type) > 0 && property.Type != "List" && property.Type != "Map",
	})
	checkError(err)
	return buf.String()
}

func needsFmtImport(cfnType CfnType) bool {
	for _, property := range cfnType.Properties {
		if len(property.PrimitiveType) > 0 {
			if property.Required {
				return true
			}
		} else if len(property.Type) > 0 {
			if property.Required {
				return true
			}
		}
	}
	return false
}

func needsPropertiesImport(cfnType CfnType) bool {
	for _, property := range cfnType.Properties {
		if len(property.Type) > 0 && property.Type != "List" && property.Type != "Map" {
			return true
		}
	}
	return false
}

func sortSpecList(specList map[string]string) []string {
	regions := make([]string, len(specList))
	i := 0
	for region := range specList {
		regions[i] = region
		i++
	}
	sort.Strings(regions)
	return regions
}

func sortProperties(properties map[string]CfnProperty) []NamedCfnProperty {
	primitives := []NamedCfnProperty{}
	nonPrimitives := []NamedCfnProperty{}
	for name, property := range properties {
		namedProperty := NamedCfnProperty{CfnProperty: property, name: name}
		if len(property.PrimitiveType) > 0 {
			primitives = append(primitives, namedProperty)
		} else {
			nonPrimitives = append(nonPrimitives, namedProperty)
		}
	}
	sort.Sort(ByName(primitives))
	sort.Sort(ByName(nonPrimitives))
	sort.Sort(ByType(nonPrimitives))
	return append(primitives, nonPrimitives...)
}

func sortAttributeNames(attributes map[string]CfnAttribute) []string {
	names := make([]string, len(attributes))
	i := 0
	for name := range attributes {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

func sortTypeNames(types map[string]CfnType) []string {
	names := make([]string, len(types))
	i := 0
	for name := range types {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

func isPropertyGlobal(typeName string) bool {
	return !strings.Contains(typeName, "::")
}

func propertyNameFromPropertyType(typeName string) string {
	if isPropertyGlobal(typeName) {
		return typeName
	}
	parts := strings.Split(typeName, "::")
	subParts := strings.Split(parts[len(parts)-1], ".")
	return strings.Join(subParts, "_")
}

func propertyNameFromResourceType(typeName, propertyName string) string {
	for _, v := range globalPropertyTypes {
		if v == propertyName {
			return propertyName
		}
	}
	parts := strings.Split(typeName, "::")
	subParts := strings.Split(parts[len(parts)-1], ".")
	return subParts[0] + "_" + propertyName
}

func fileNameFromCfnType(typeName string) string {
	parts := strings.Split(typeName, "::")
	return fmt.Sprint(strings.ToLower(parts[1]), "_", strings.ToLower(parts[2]))
}

func titleCaseNameFromCfnType(typeName string) string {
	parts := strings.Split(typeName, "::")
	return fmt.Sprint(strings.Title(parts[1]), strings.Title(parts[2]))
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Sort types
type ByType []NamedCfnProperty

func (a ByType) Len() int      { return len(a) }
func (a ByType) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByType) Less(i, j int) bool {
	return a[i].Type > a[j].Type
}

type ByName []NamedCfnProperty

func (a ByName) Len() int      { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool {
	return a[i].name < a[j].name
}
