package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
)

// Generate the types.go file.
func (data *Data) generateTypes(doc *openapi3.T) error {
	// Iterate over all the schema components in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Components.Schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		s := doc.Components.Schemas[name]
		if s.Ref != "" {
			// We are going through all the reference, so we will catch this one.
			continue
		}

		if err := data.generateSchemaType(name, s.Value); err != nil {
			return err
		}
	}

	// Iterate over all the responses in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	rKeys := make([]string, 0)
	for k := range doc.Components.Responses {
		rKeys = append(rKeys, k)
	}
	sort.Strings(rKeys)
	for _, name := range rKeys {
		r := doc.Components.Responses[name]
		if r.Ref != "" {
			logrus.Warnf("TODO: skipping response for %q, since it is a reference", name)
			continue
		}

		if err := data.generateResponseType(name, r.Value); err != nil {
			return err
		}
	}

	// TODO: write the types for the parameters if defined in components.

	// Write the types to the template.
	if err := processTemplate("types.tmpl", "types.go", *data); err != nil {
		return err
	}

	return nil
}

// generateSchemaType writes a type definition for the given schema.
// The additional parameter is only used as a suffix for the type name.
// This is mostly for oneOf types.
func (data *Data) generateSchemaType(name string, s *openapi3.Schema) error {
	otype := s.Type
	logrus.Debugf("writing type for schema %q -> %s", name, otype)

	name = printProperty(name)

	if otype == "string" {
		// If this is an enum, write the enum type.
		if len(s.Enum) > 0 {
			if err := data.generateEnumType(name, s); err != nil {
				return err
			}
		} else {
			// TODO: fmt.Fprintf(f, "type %s string\n", name)
		}
	} else if otype == "integer" {
		//TODO	fmt.Fprintf(f, "type %s int\n", name)
	} else if otype == "number" {
		//TODO	fmt.Fprintf(f, "type %s float64\n", name)
	} else if otype == "boolean" {
		//TODO	fmt.Fprintf(f, "type %s bool\n", name)
	} else if otype == "array" {
		// TODO	fmt.Fprintf(f, "type %s []%s\n", name, s.Items.Value.Type)
	} else if otype == "object" {
		if err := data.generateObjectType(name, s); err != nil {
			return err
		}
	} else if s.OneOf != nil {
		if err := data.generateOneOfType(name, s); err != nil {
			return err
		}
	} else if s.AnyOf != nil {
		logrus.Warnf("TODO: skipping type for %q, since it is a ANYOF", name)
	} else if s.AllOf != nil {
		logrus.Warnf("TODO: skipping type for %q, since it is a ALLOF", name)
	}

	return nil
}

// generateResponseType writes a type definition for the given response.
func (data *Data) generateResponseType(name string, r *openapi3.Response) error {
	// Write the type definition.
	for k, v := range r.Content {
		logrus.Debugf("writing type for response %q -> %q", name, k)

		name := fmt.Sprintf("Response%s", name)

		// Write the type description.
		// TODO: fix
		/*if r.Description != nil {
			fmt.Fprintf(f, "// %s is the response given when %s\n", name, toLowerFirstLetter(
				strings.ReplaceAll(*r.Description, "\n", "\n// ")))
		} else {
			fmt.Fprintf(f, "// %s is the type definition for a %s response.\n", name, name)
		}*/

		if err := data.generateSchemaType(name, v.Schema.Value); err != nil {
			return err
		}
	}

	return nil
}

// Enum holds the information for an enum.
type Enum struct {
	Name        string
	Description string
	Values      []EnumValue
}

// EnumValue holds the information for an enum value.
type EnumValue struct {
	Name  string
	Value string
}

func (data *Data) generateEnumType(name string, s *openapi3.Schema) error {
	enumName := makeSingular(name)
	enum := Enum{
		Name:        enumName,
		Description: getTypeDescription(enumName, s),
		Values:      []EnumValue{},
	}

	for _, v := range s.Enum {
		// Most likely, the enum values are strings.
		enumValue, ok := v.(string)
		if !ok {
			return fmt.Errorf("enum value for %q is not a string: %#v", name, v)
		}
		enumValue = strings.TrimSpace(enumValue)

		enumValueName := printProperty(fmt.Sprintf("%s %s", enumName, enumValue))

		if enumValue == "" {
			// Make the type have empty in the name.
			enumValueName = fmt.Sprintf("%sEmpty", enumValueName)
		}

		enum.Values = append(enum.Values, EnumValue{
			Name:  enumValueName,
			Value: enumValue,
		})
	}

	// Print the template for the enum.
	enumString, err := templateToString("enum.tmpl", enum)
	if err != nil {
		return err
	}

	// Add the type to our types.
	data.Types = append(data.Types, enumString)

	return nil
}

// Object holds the information for an object.
type Object struct {
	Name        string
	Description string
	Values      map[string]ObjectValue
}

// ObjectValue holds the information for an object value.
type ObjectValue struct {
	Name        string
	Description string
	Type        string
	Property    string
}

func (data *Data) generateObjectType(name string, s *openapi3.Schema) error {
	object := Object{
		Name:        name,
		Description: getTypeDescription(name, s),
		Values:      map[string]ObjectValue{},
	}

	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range s.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := s.Properties[k]

		objectValue := ObjectValue{
			Name:     printProperty(k),
			Type:     printType(k, v),
			Property: k,
		}

		if v.Value.Description != "" {
			objectValue.Description = strings.ReplaceAll(v.Value.Description, "\n", "\n// ")
		}

		object.Values[k] = objectValue
	}

	// Print the template for the struct.
	objectString, err := templateToString("struct.tmpl", object)
	if err != nil {
		return err
	}

	// Add the type to our types.
	data.Types = append(data.Types, objectString)

	return nil
}

func (data *Data) generateOneOfType(name string, s *openapi3.Schema) error {
	logrus.Warnf("TODO: oneof type for %q", name)
	return nil
}

func getReferenceSchema(v *openapi3.SchemaRef) string {
	if v.Ref != "" {
		ref := strings.TrimPrefix(v.Ref, "#/components/schemas/")
		if len(v.Value.Enum) > 0 {
			return printProperty(makeSingular(ref))
		}

		return printProperty(ref)
	}

	return ""
}

// getTypeDescription gets the description of the given type.
func getTypeDescription(name string, s *openapi3.Schema) string {
	if s.Description != "" {
		return fmt.Sprintf("%s: %s", name, strings.ReplaceAll(s.Description, "\n", "\n// "))
	}

	return fmt.Sprintf("%s is the type definition for a %s.", name, name)
}

func isLocalEnum(v *openapi3.SchemaRef) bool {
	return v.Ref == "" && v.Value.Type == "string" && len(v.Value.Enum) > 0
}

func isLocalObject(v *openapi3.SchemaRef) bool {
	return v.Ref == "" && v.Value.Type == "object" && len(v.Value.Properties) > 0
}

// formatStringType converts a string schema to a valid Go type.
func formatStringType(t *openapi3.Schema) string {
	if t.Format == "date-time" {
		return "*JSONTime"
	} else if t.Format == "partial-date-time" {
		return "*JSONTime"
	} else if t.Format == "date" {
		return "*JSONTime"
	} else if t.Format == "time" {
		return "*JSONTime"
	} else if t.Format == "email" {
		return "string"
	} else if t.Format == "hostname" {
		return "string"
	} else if t.Format == "ip" || t.Format == "ipv4" || t.Format == "ipv6" {
		return "net.IP"
	} else if t.Format == "uri" || t.Format == "url" {
		return "url.URL"
	} else if t.Format == "uuid" {
		return "uuid.UUID"
	} else if t.Format == "uuid3" {
		return "string"
	}

	return "string"
}

// printType converts a schema type to a valid Go type.
func printType(property string, r *openapi3.SchemaRef, spec *openapi3.T) (string, error) {
	s := r.Value
	t := s.Type

	// If we have a reference, we can usually just use that.
	if r.Ref != "" {
		// Get the schema for the reference.
		ref := strings.TrimPrefix(r.Ref, "#/components/schemas/")
		reference, ok := spec.Components.Schemas[ref]
		if !ok {
			return "", fmt.Errorf("reference %q not found in schemas", ref)
		}

		// If the reference is an object or an enum, return the reference.
		if reference.Value.Type == "object" || reference.Value.Type == "string" && len(reference.Value.Enum) > 0 {
			return getReferenceSchema(r), nil
		}

		// Otherwise, we need to recurse.
		return printType(property, reference, spec)
	}

	// See if we have an allOf.
	if s.AllOf != nil {
		if len(s.AllOf) > 1 {
			return "", fmt.Errorf("allOf for %q has more than 1 item", property)
		}

		return printType(property, s.AllOf[0]), nil
	}

	if t == "string" {
		reference := getReferenceSchema(r)
		if reference != "" {
			return reference
		}

		return formatStringType(s), nil
	} else if t == "integer" {
		return "int"
	} else if t == "number" {
		return "float64"
	} else if t == "boolean" {
		return "bool"
	} else if t == "array" {
		reference := getReferenceSchema(s.Items)
		if reference != "" {
			return fmt.Sprintf("[]%s", reference), nil
		}

		// TODO: handle if it is not a reference.
		return "[]string"
	} else if t == "object" {
		if s.AdditionalProperties != nil {
			return printType(property, s.AdditionalProperties)
		}
		// Most likely this is a local object, we will handle it.
		return strcase.ToCamel(property), nil
	}

	return "", fmt.Errorf("unknown type %q for %q", t, property)
}
