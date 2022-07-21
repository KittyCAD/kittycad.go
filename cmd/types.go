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

		if err := data.generateSchemaType(name, s.Value, doc); err != nil {
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

		if err := data.generateResponseType(name, r.Value, doc); err != nil {
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
func (data *Data) generateSchemaType(name string, s *openapi3.Schema, spec *openapi3.T) error {
	otype := s.Type
	logrus.Debugf("writing type for schema %q -> %s", name, otype)

	name = printProperty(name)

	if otype == "string" {
		// If this is an enum, write the enum type.
		if len(s.Enum) > 0 {
			if err := data.generateEnumType(name, s); err != nil {
				return err
			}
		}
	} else if otype == "object" {
		if err := data.generateObjectType(name, s, spec); err != nil {
			return err
		}
	} else if s.OneOf != nil {
		if err := data.generateOneOfType(name, s, spec); err != nil {
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
func (data *Data) generateResponseType(name string, r *openapi3.Response, spec *openapi3.T) error {
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

		if err := data.generateSchemaType(name, v.Schema.Value, spec); err != nil {
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

func (data *Data) generateObjectType(name string, s *openapi3.Schema, spec *openapi3.T) error {
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

		typeName, err := printType(k, v, spec)
		if err != nil {
			return err
		}

		objectValue := ObjectValue{
			Name:     printProperty(k),
			Type:     typeName,
			Property: k,
		}

		if v.Value.Description != "" {
			objectValue.Description = strings.ReplaceAll(v.Value.Description, "\n", "\n// ")
		} else {
			// Try another way.
			description, err := getDescriptionForSchemaOrReference(v, spec)
			if err != nil {
				return err
			}
			if description != "" {
				objectValue.Description = strings.ReplaceAll(description, "\n", "\n// ")
			}
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

func (data *Data) generateOneOfType(name string, s *openapi3.Schema, spec *openapi3.T) error {
	logrus.Warnf("TODO: oneof type for %q", name)
	return data.generateObjectType(name, s.OneOf[0].Value, spec)
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

// formatStringType converts a string schema to a valid Go type.
func formatStringType(t *openapi3.Schema) string {
	if t.Format == "date-time" {
		return "Time"
	} else if t.Format == "partial-date-time" {
		return "Time"
	} else if t.Format == "date" {
		return "Time"
	} else if t.Format == "time" {
		return "Time"
	} else if t.Format == "email" {
		return "string"
	} else if t.Format == "hostname" {
		return "string"
	} else if t.Format == "ip" || t.Format == "ipv4" || t.Format == "ipv6" {
		return "IP"
	} else if t.Format == "byte" {
		return "Base64"
	} else if t.Format == "uri" || t.Format == "url" {
		return "URL"
	} else if t.Format == "uuid" {
		return "UUID"
	} else if t.Format == "uuid3" {
		return "string"
	} else if t.Format == "binary" {
		return "[]byte"
	}

	return "string"
}

func isTypeToString(s string) bool {
	s = strings.TrimPrefix(s, "*")
	return s == "URL" || s == "UUID" || s == "IP" || s == "Time"
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

		return printType(property, s.AllOf[0], spec)
	}

	if t == "string" {
		reference := getReferenceSchema(r)
		if reference != "" {
			return reference, nil
		}

		return formatStringType(s), nil
	} else if t == "integer" {
		return "int", nil
	} else if t == "number" {
		return "float64", nil
	} else if t == "boolean" {
		return "bool", nil
	} else if t == "array" {
		reference := getReferenceSchema(s.Items)
		if reference != "" {
			return fmt.Sprintf("[]%s", reference), nil
		}

		// TODO: handle if it is not a reference.
		return "[]string", nil
	} else if t == "object" {
		if s.AdditionalProperties != nil {
			return printType(property, s.AdditionalProperties, spec)
		}
		// Most likely this is a local object, we will handle it.
		return strcase.ToCamel(property), nil
	}

	return "interface{}", nil
}

func printTagName(tag string) string {
	return printProperty(makeSingular(tag))
}

// printProperty converts an object's property name to a valid Go identifier.
func printProperty(p string) string {
	c := strcase.ToCamel(p)
	if c == "Id" {
		c = "ID"
	} else if c == "IpAddress" {
		c = "IPAddress"
	} else if c == "UserId" {
		c = "UserID"
	} else if strings.HasPrefix(c, "Gpu") {
		c = strings.Replace(c, "Gpu", "GPU", 1)
	} else if strings.HasSuffix(c, "Id") {
		c = strings.TrimSuffix(c, "Id") + "ID"
	}

	c = strings.ReplaceAll(c, "Api", "API")

	return c
}

func printPropertyLower(p string) string {
	s := strcase.ToLowerCamel(printProperty(p))

	if s == "iD" {
		s = "id"
	} else if s == "iPAddress" {
		s = "ipAddress"
	} else if s == "iDSortMode" {
		s = "idSortMode"
	}

	return s
}

// makeSingular returns the given string but singular.
func makeSingular(s string) string {
	if strings.HasSuffix(s, "Status") {
		return s
	}
	return strings.TrimSuffix(s, "s")
}
