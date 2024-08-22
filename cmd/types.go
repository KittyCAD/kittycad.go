package main

import (
	"fmt"
	"reflect"
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
			if err := data.generateEnumType(name, s, map[string]string{}); err != nil {
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
	Name        string
	Description string
	Value       string
}

func (data *Data) generateEnumType(name string, s *openapi3.Schema, additionalDocs map[string]string) error {
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

		enumValueStruct := EnumValue{
			Name:  enumValueName,
			Value: enumValue,
		}

		if docs, ok := additionalDocs[enumValue]; ok {
			if docs != "" {
				enumValueStruct.Description = fmt.Sprintf("%s: %s", enumValueName, strings.ReplaceAll(docs, "\n", "\n// "))
			}
		}

		enum.Values = append(enum.Values, enumValueStruct)
	}

	// Print the template for the enum.
	enumString, err := templateToString("enum.tmpl", enum)
	if err != nil {
		return err
	}

	// Add the type to our types.
	data.Types[enum.Name] = enumString

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
	Required    bool
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
			Required: contains(s.Required, k),
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

		if v.Value.Deprecated {
			objectValue.Description += "\n//\n// Deprecated: " + printProperty(k) + " is deprecated."

		}

		object.Values[k] = objectValue

		// If this property is an object, we need to generate it as well.
		if v.Value.Type == "object" && v.Value.Properties != nil && len(v.Value.Properties) > 0 {
			// Check if we already have a schema for this one of.
			if _, ok := spec.Components.Schemas[k]; !ok {
				n := printProperty(k)
				if n == "Error" {
					n = name + "Error"
				}
				if err := data.generateObjectType(n, v.Value, spec); err != nil {
					return err
				}
			}
		}
	}

	// Print the template for the struct.
	objectString, err := templateToString("struct.tmpl", object)
	if err != nil {
		return err
	}

	// Add the type to our types.
	data.Types[object.Name] = objectString

	return nil
}

func (data *Data) generateOneOfType(name string, s *openapi3.Schema, spec *openapi3.T) error {
	// Check if this is an enum with descriptions.
	isEnumWithDocs := false
	enumDocs := map[string]string{}
	enumeration := []interface{}{}
	for _, oneOf := range s.OneOf {
		if oneOf.Value.Type == "string" && oneOf.Value.Enum != nil && len(oneOf.Value.Enum) == 1 {
			// Get the description for this enum.
			isEnumWithDocs = true
			enumDocs[oneOf.Value.Enum[0].(string)] = oneOf.Value.Description
			enumeration = append(enumeration, oneOf.Value.Enum[0])
		} else {
			isEnumWithDocs = false
			break
		}
	}

	if isEnumWithDocs {
		return data.generateEnumType(name, &openapi3.Schema{
			Type:        "string",
			Description: s.Description,
			Enum:        enumeration,
		}, enumDocs)
	}

	if len(s.OneOf) == 1 && s.OneOf[0].Value.Type == "object" {
		// We need to generate the one of type.
		return data.generateSchemaType(name, s.OneOf[0].Value, spec)
	}

	// Check if they all have a type.
	types := []string{}
	typeName := ""
	for _, v := range s.OneOf {
		if v.Value.Type == "object" {
			keys := []string{}
			for k := range v.Value.Properties {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, propName := range keys {
				value := v.Value.Properties[propName]
				// Check if all the objects have a enum of one type.
				if value.Value.Type == "string" && value.Value.Enum != nil && len(value.Value.Enum) == 1 {
					if typeName == "" {
						typeName = propName
					} else if typeName != propName {
						return fmt.Errorf("one of %q has a different type than the others: %q", name, value.Value.Enum[0].(string))
					}
					types = append(types, name+" "+value.Value.Enum[0].(string))
				} else {
					types = append(types, name+" "+propName)
				}
			}
		} else if v.Value.Type == "string" && v.Value.Enum != nil && len(v.Value.Enum) == 1 {
			types = append(types, v.Value.Enum[0].(string))
		}
	}

	for index, oneOf := range s.OneOf {
		// Check if we already have this type defined.
		iname := printProperty(types[index])
		if _, ok := data.Types[iname]; ok {
			// We should name the type after the one of.
			iname = printProperty(name + " " + types[index])
		}

		// Check if we already have a schema for this one of.
		reference, ok := spec.Components.Schemas[types[index]]
		if !ok {
			if err := data.generateSchemaType(iname, oneOf.Value, spec); err != nil {
				return err
			}
		}

		// Remove the type from the properties.
		properties := oneOf.Value.Properties
		delete(properties, typeName)

		// Make sure they are equal.
		if reference != nil && reference.Value != nil && reference.Value.Properties != nil && properties != nil && reflect.DeepEqual(reference.Value.Properties, properties) {
			// We need to generate the one of type.
			if err := data.generateSchemaType(fmt.Sprintf("%s %s", name, types[index]), oneOf.Value, spec); err != nil {
				return err
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

func printOneOf(property string, r *openapi3.SchemaRef, spec *openapi3.T) (string, error) {
	s := r.Value

	// Check if this is an enum with descriptions.
	isEnumWithDocs := false
	enumeration := []interface{}{}
	for _, oneOf := range s.OneOf {
		if oneOf.Value.Type == "string" && oneOf.Value.Enum != nil && len(oneOf.Value.Enum) == 1 {
			// Get the description for this enum.
			isEnumWithDocs = true
			enumeration = append(enumeration, oneOf.Value.Enum[0])
		} else {
			isEnumWithDocs = false
			break
		}
	}

	if isEnumWithDocs {
		newSchema := &openapi3.SchemaRef{
			Ref: r.Ref,
			Value: &openapi3.Schema{
				Type:        "string",
				Description: s.Description,
				Enum:        enumeration,
			}}

		if r.Ref != "" {
			return getReferenceSchema(newSchema), nil
		}
		return printType(property, newSchema, spec)
	}

	return "any", nil
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
		// If we have a oneOf we are going to use a generic for it.
		if reference.Value.OneOf != nil {
			return printOneOf(property, r, spec)
		} else if reference.Value.Type == "object" || reference.Value.Type == "string" && len(reference.Value.Enum) > 0 {
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

	if s.OneOf != nil {
		return printOneOf(property, r, spec)
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

		// Get the type for the schema.
		innerType, err := printType(property, s.Items, spec)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("[]%s", innerType), nil
	} else if t == "object" {
		if s.AdditionalProperties.Schema != nil && (s.Properties == nil || len(s.Properties) == 0) {
			// get the inner type.
			innerType, err := printType(property, s.AdditionalProperties.Schema, spec)
			if err != nil {
				return "", err
			}
			// Now make it a map.
			return fmt.Sprintf("map[string]%s", innerType), nil
		}
		// Most likely this is a local object, we will handle it.
		return strcase.ToCamel(property), nil
	}

	return "any", nil
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
	c = strings.ReplaceAll(c, "APIcall", "APICall")
	c = strings.ReplaceAll(c, "APItoken", "APIToken")

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

// Check if a slice of strings contains a value.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Generate a random value based on a schema.
func (data Data) generateExampleValue(name string, s *openapi3.SchemaRef, spec *openapi3.T, required bool) (string, error) {
	schema := s.Value
	typet := s.Value.Type

	// If we have a reference, we can usually just use that.
	if s.Ref != "" {
		// Get the schema for the reference.
		ref := strings.TrimPrefix(s.Ref, "#/components/schemas/")
		reference, ok := spec.Components.Schemas[ref]
		if !ok {
			return "", fmt.Errorf("reference %q not found in schemas", ref)
		}

		// If the reference is an object or an enum, return the reference.
		// Otherwise, we need to recurse.
		return data.generateExampleValue(printProperty(ref), reference, spec, true)
	}

	if typet == "string" {
		// Check if we have an enum.
		if len(schema.Enum) > 0 {
			// Get the first enum value.
			firstValue := printProperty(schema.Enum[0].(string))
			t := fmt.Sprintf("%s.%s%s", data.PackageName, name, firstValue)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		}

		if schema.Format == "date-time" {
			t := fmt.Sprintf(`%s.TimeNow()`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "partial-date-time" {
			t := fmt.Sprintf(`%s.TimeNow()`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "date" {
			t := fmt.Sprintf(`%s.TimeNow()`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "time" {
			t := fmt.Sprintf(`%s.TimeNow()`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "email" {
			return `"example@example.com"`, nil
		} else if schema.Format == "hostname" {
			return `"localhost"`, nil
		} else if schema.Format == "ip" || schema.Format == "ipv4" || schema.Format == "ipv6" {
			t := fmt.Sprintf(`%s.IP{netaddr.MustParseIP("192.158.1.38")}`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "byte" {
			t := fmt.Sprintf(`%s.Base64{Inner: []byte("aGVsbG8gd29ybGQK")}`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "uri" || schema.Format == "url" {
			t := fmt.Sprintf(`%s.URL{&url.URL{Scheme: "https", Host: "example.com"}}`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "uuid" {
			t := fmt.Sprintf(`%s.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8")`, data.PackageName)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "binary" {
			t := `[]byte("some-binary")`
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		} else if schema.Format == "phone" {
			return `"+1-555-555-555"`, nil
		}

		return `"some-string"`, nil
	} else if typet == "integer" {
		return "123", nil
	} else if typet == "number" {
		return "123.45", nil
	} else if typet == "boolean" {
		return "true", nil
	} else if typet == "array" {
		// Get the type name.
		typeName, err := printType(name, schema.Items, spec)
		if err != nil {
			return "", err
		}

		// Get an example for the items.
		items, err := data.generateExampleValue(name, schema.Items, spec, true)
		if err != nil {
			return "", err
		}

		if schema.Items.Ref != "" {
			t := fmt.Sprintf("[]kittycad.%s{%s}", typeName, items)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		}

		t := fmt.Sprintf("[]%s{%s}", typeName, items)
		if required {
			return t, nil
		}
		return fmt.Sprintf("&%s", t), nil
	} else if typet == "object" {
		if schema.AdditionalProperties.Schema != nil && (schema.Properties == nil || len(schema.Properties) == 0) {
			// get the inner type.
			innerType, err := printType(name, schema.AdditionalProperties.Schema, spec)
			if err != nil {
				return "", err
			}

			if schema.AdditionalProperties.Schema.Value.Type == "object" {
				innerType = fmt.Sprintf("%s.%s", "kittycad", innerType)
			}

			// Get an example for the inner type.
			innerExample, err := data.generateExampleValue(name, schema.AdditionalProperties.Schema, spec, true)
			if err != nil {
				return "", err
			}

			// Now make it a map.
			t := fmt.Sprintf("map[string]%s{\"example\": %s}", innerType, innerExample)
			if required {
				return t, nil
			}
			return fmt.Sprintf("&%s", t), nil
		}

		// Get the type name.
		typeName, err := printType(name, s, spec)
		if err != nil {
			return "", err
		}

		object := fmt.Sprintf("%s.%s{", data.PackageName, typeName)
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range schema.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			p := schema.Properties[k]
			// Get an example for the property.
			example, err := data.generateExampleValue(printProperty(k), p, spec, required)
			if err != nil {
				return "", err
			}
			object += fmt.Sprintf("%s: %s, ", printProperty(k), example)
		}
		// Close the object.
		object += "}"

		if required {
			return object, nil
		}
		return fmt.Sprintf("&%s", object), nil
	}

	if schema.AllOf != nil {
		return data.generateExampleValue(name, schema.AllOf[0], spec, required)
	}

	return `""`, nil
}
