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

		if err := data.generateSchemaType(name, s.Value, ""); err != nil {
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
func (data *Data) generateSchemaType(name string, s *openapi3.Schema, additionalName string) error {
	otype := s.Type
	logrus.Debugf("writing type for schema %q -> %s", name, otype)

	name = printProperty(name)
	typeName := strings.ReplaceAll(strings.TrimSpace(fmt.Sprintf("%s%s", name, printProperty(additionalName))), "Api", "API")

	if otype == "string" {
		// If this is an enum, write the enum type.
		if len(s.Enum) > 0 {
			if err := data.generateEnumType(typeName, s); err != nil {
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
		if err := data.generateObjectType(typeName, s); err != nil {
			return err
		}
	} else if s.OneOf != nil {
		// We want to convert these to a different data type to be more idiomatic.
		// But first, we need to make sure we have a type for each one.
		var oneOfTypes []string
		var properties []string
		for _, v := range s.OneOf {
			// We want to iterate over the properties of the embedded object
			// and find the type that is a string.
			var typeName string

			// Iterate over all the schema components in the spec and write the types.
			// We want to ensure we keep the order so the diffs don't look like shit.
			keys := make([]string, 0)
			for k := range v.Value.Properties {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, prop := range keys {
				p := v.Value.Properties[prop]
				// We want to collect all the unique properties to create our global oneOf type.
				propertyName := printType(prop, p)

				propertyString := fmt.Sprintf("\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", printProperty(prop), propertyName, prop, prop)
				if !containsMatchFirstWord(properties, propertyString) {
					properties = append(properties, propertyString)
				}

				if p.Value.Type == "string" {
					if p.Value.Enum != nil {
						// We want to get the enum value.
						// Make sure there is only one.
						if len(p.Value.Enum) != 1 {
							logrus.Warnf("TODO: oneOf for %q -> %q enum %#v", name, prop, p.Value.Enum)
							continue
						}

						typeName = printProperty(p.Value.Enum[0].(string))
					}
				}
			}

			// Basically all of these will have one type embedded in them that is a
			// string and the type, since these come from a Rust sum type.
			oneOfType := fmt.Sprintf("%s%s", name, typeName)
			generateSchemaType(f, name, v.Value, typeName)
			// Add it to our array.
			oneOfTypes = append(oneOfTypes, oneOfType)
		}

		// Now let's create the global oneOf type.
		// Write the type description.
		getTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s struct {\n", typeName)
		// Iterate over the properties and write the types, if we need to.
		for _, p := range properties {
			fmt.Fprintf(f, p)
		}
		// Close the struct.
		fmt.Fprintf(f, "}\n")

	} else if s.AnyOf != nil {
		logrus.Warnf("TODO: skipping type for %q, since it is a ANYOF", name)
	} else if s.AllOf != nil {
		logrus.Warnf("TODO: skipping type for %q, since it is a ALLOF", name)
	}

	// Add a newline at the end of the type.
	fmt.Fprintln(f, "")

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

		if err := data.generateSchemaType(f, name, v.Schema.Value, ""); err != nil {
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
		enum, ok := v.(string)
		if !ok {
			return fmt.Errorf("enum value for %q is not a string: %#v", name, v)
		}

		enum.Values = append(enum.Values, EnumValue{
			Name:  strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enumName)),
			Value: enum,
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

func (data *Data) generateObjectType(name string, s *openapi3.Schema) error {
	// Get the type description.
	description := getTypeDescription(typeName, s)
	recursive := false
	fmt.Fprintf(f, "type %s struct {\n", typeName)
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range s.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := s.Properties[k]
		// Check if we need to generate a type for this type.
		typeName := printType(k, v)

		if isLocalEnum(v) {
			recursive = true
			typeName = fmt.Sprintf("%s%s", name, printProperty(k))
		}

		if isLocalObject(v) {
			recursive = true
			logrus.Warnf("TODO: skipping object for %q -> %#v", name, v)
			typeName = fmt.Sprintf("%s%s", name, printProperty(k))
		}

		if v.Value.Description != "" {
			fmt.Fprintf(f, "\t// %s is %s\n", printProperty(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
		}
		fmt.Fprintf(f, "\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", printProperty(k), typeName, k, k)
	}
	fmt.Fprintf(f, "}\n")

	if recursive {
		// Add a newline at the end of the type.
		fmt.Fprintln(f, "")

		// Iterate over the properties and write the types, if we need to.
		for k, v := range s.Properties {
			if isLocalEnum(v) {
				generateSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
			}

			if isLocalObject(v) {
				generateSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
			}
		}
	}

	return nil
}

func getReferenceSchema(v *openapi3.SchemaRef) string {
	if v.Ref != "" {
		ref := strings.TrimPrefix(v.Ref, "#/components/schemas/")
		if len(v.Value.Enum) > 0 {
			return printProperty(makeSingular(ref))
		}

		return strings.ReplaceAll(printProperty(ref), "Api", "API")
	}

	return ""
}

// getTypeDescription gets the description of the given type.
func getTypeDescription(name string, s *openapi3.Schema) string {
	if s.Description != "" {
		return fmt.Sprintf("%s: %s", name, s.Description)
	}

	fmt.Sprintf("%s is the type definition for a %s.", name, name)
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
	} else if t.Format == "ipv4" {
		return "string"
	} else if t.Format == "ipv6" {
		return "string"
	} else if t.Format == "uri" {
		return "string"
	} else if t.Format == "uuid" {
		return "string"
	} else if t.Format == "uuid3" {
		return "string"
	}

	return "string"
}

// printType converts a schema type to a valid Go type.
func printType(property string, r *openapi3.SchemaRef) string {
	s := r.Value
	t := s.Type

	// If we have a reference, just use that.
	if r.Ref != "" {
		return getReferenceSchema(r)
	}

	// See if we have an allOf.
	if s.AllOf != nil {
		if len(s.AllOf) > 1 {
			logrus.Warnf("TODO: allOf for %q has more than 1 item", property)
			return "TODO"
		}

		return printType(property, s.AllOf[0])
	}

	if t == "string" {
		reference := getReferenceSchema(r)
		if reference != "" {
			return reference
		}

		return formatStringType(s)
	} else if t == "integer" {
		return "int"
	} else if t == "number" {
		return "float64"
	} else if t == "boolean" {
		return "bool"
	} else if t == "array" {
		reference := getReferenceSchema(s.Items)
		if reference != "" {
			return fmt.Sprintf("[]%s", reference)
		}

		// TODO: handle if it is not a reference.
		return "[]string"
	} else if t == "object" {
		if s.AdditionalProperties != nil {
			return printType(property, s.AdditionalProperties)
		}
		// Most likely this is a local object, we will handle it.
		return strcase.ToCamel(property)
	}

	logrus.Warnf("TODO: skipping type %q for %q, marking as interface{}", t, property)
	return "interface{}"
}
