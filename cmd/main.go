package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/jsondiff"
)

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	// Load the open API spec from the file.
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}
	p := filepath.Join(wd, "spec.json")

	doc, err := openapi3.NewLoader().LoadFromFile(p)
	if err != nil {
		return fmt.Errorf("error loading openAPI spec: %v", err)
	}

	data := Data{
		PackageName:      "kittycad",
		BaseURL:          "https://api.kittycad.io",
		EnvVariable:      "KITTYCAD_API_TOKEN",
		Tags:             []Tag{},
		WorkingDirectory: wd,
		Examples:         []string{},
		Paths:            []string{},
		Types:            []string{},
	}
	// Format the tags for our data.
	for _, tag := range doc.Tags {
		data.Tags = append(data.Tags, Tag{
			Name:        printTagName(tag.Name),
			Description: tag.Description,
		})
	}

	// Render the client examples.
	clientInfo, err := templateToString("client-example.tmpl", data)
	if err != nil {
		return fmt.Errorf("error processing template: %v", err)
	}
	data.Examples = append(data.Examples, clientInfo)
	doc.Info.Extensions["x-go"] = map[string]string{
		"install": "go get github.com/kittycad/kittycad.go",
		"client":  clientInfo,
	}

	// Generate the client.go file.
	logrus.Info("Generating client...")
	if err := generateClient(doc, data); err != nil {
		return err
	}

	// Generate the types.go file.
	logrus.Info("Generating types...")
	generateTypes(doc)

	// Generate the responses.go file.
	logrus.Info("Generating responses...")
	generateResponses(doc)

	// Generate the paths.go file.
	logrus.Info("Generating paths...")
	if err := data.generatePaths(doc); err != nil {
		return err
	}

	// Generate the examples.go file.
	logrus.Info("Generating examples...")
	if err := generateExamplesFile(doc, data); err != nil {
		return err
	}

	// Get the old doc again.
	oldDoc, err := openapi3.NewLoader().LoadFromFile(p)
	if err != nil {
		return fmt.Errorf("error loading openAPI spec: %v", err)
	}
	patch, err := jsondiff.Compare(oldDoc, doc)
	if err != nil {
		logrus.Errorf("error comparing old and new openAPI spec: %v", err)
	}
	patchJSON, err := json.MarshalIndent(patch, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling openAPI spec: %v", err)
	}

	diffFile := filepath.Join(wd, "kittycad.go.patch.json")
	if err := ioutil.WriteFile(diffFile, patchJSON, 0644); err != nil {
		return fmt.Errorf("error writing openAPI spec patch to %s: %v", diffFile, err)
	}

	return nil
}

var enumStringTypes map[string][]string = map[string][]string{}

// Generate the types.go file.
func generateTypes(doc *openapi3.T) {
	f := openGeneratedFile("types.go")
	defer f.Close()

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
			logrus.Warnf("TODO: skipping type for %q, since it is a reference", name)
			continue
		}

		writeSchemaType(f, name, s.Value, "")
	}

	// Iterate over all the enum types and add in the slices.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys = make([]string, 0)
	for k := range enumStringTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		enums := enumStringTypes[name]
		// Make the enum a collection of the values.
		// Add a description.
		fmt.Fprintf(f, "// %s is the collection of all %s values.\n", makePlural(name), makeSingular(name))
		fmt.Fprintf(f, "var %s = []%s{\n", makePlural(name), makeSingular(name))
		// We want to keep the values in the same order as the enum.
		sort.Strings(enums)
		for _, enum := range enums {
			// Most likely, the enum values are strings.
			fmt.Fprintf(f, "\t%s,\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)))
		}
		// Close the enum values.
		fmt.Fprintf(f, "}\n")
	}
}

// Generate the responses.go file.
func generateResponses(doc *openapi3.T) {
	f := openGeneratedFile("responses.go")
	defer f.Close()

	// Iterate over all the responses in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Components.Responses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		r := doc.Components.Responses[name]
		if r.Ref != "" {
			logrus.Warnf("TODO: skipping response for %q, since it is a reference", name)
			continue
		}

		writeResponseType(f, name, r.Value)
	}
}

// Generate the client.go file.
func generateClient(doc *openapi3.T, data Data) error {
	// Generate the lib template.
	if err := processTemplate("lib.tmpl", "lib.go", data); err != nil {
		return err
	}

	// Generate the client template.
	if err := processTemplate("client.tmpl", "client.go", data); err != nil {
		return err
	}

	return nil
}

func generateExamplesFile(doc *openapi3.T, data Data) error {
	// Generate the example template.
	// All examples lack output because:
	// Examples without output comments are useful for demonstrating code that cannot run as unit tests, such as that which accesses the network, while guaranteeing the example at least compiles. (https://go.dev/blog/examples)
	// If we executed the examples it might delete a user in production or something.
	if err := processTemplate("examples.tmpl", "examples_test.go", data); err != nil {
		return err
	}

	return nil
}

func printTagName(tag string) string {
	return strings.ReplaceAll(strcase.ToCamel(makeSingular(tag)), "Api", "API")
}

func openGeneratedFile(filename string) *os.File {
	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	p := filepath.Join(cwd, filename)

	// Create the types.go file.
	// Open the file for writing.
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("error creating %q: %v\n", p, err)
		os.Exit(1)
	}

	// Add the header to the package.
	fmt.Fprintf(f, "// Code generated by `%s`. DO NOT EDIT.\n\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(f, "package kittycad")
	fmt.Fprintln(f, "")

	return f
}

func cleanFnName(name string, tag string, path string) string {
	name = printProperty(name)

	if strings.HasSuffix(tag, "s") {
		tag = strings.TrimSuffix(tag, "s")
	}

	snake := strcase.ToSnake(name)
	snake = strings.ReplaceAll(snake, "_"+strings.ToLower(tag)+"_", "_")

	name = strcase.ToCamel(snake)

	name = strings.ReplaceAll(name, "Api", "API")
	name = strings.ReplaceAll(name, "Gpu", "GPU")

	if strings.HasSuffix(name, "Get") && !strings.HasSuffix(path, "}") {
		name = fmt.Sprintf("%sList", strings.TrimSuffix(name, "Get"))
	}

	if strings.HasSuffix(name, "Post") {
		name = fmt.Sprintf("%sCreate", strings.TrimSuffix(name, "Post"))
	}

	if strings.HasPrefix(name, "s") {
		name = strings.TrimPrefix(name, "s")
	}

	if strings.Contains(name, printTagName(tag)) {
		name = strings.ReplaceAll(name, printTagName(tag)+"s", "")
		name = strings.ReplaceAll(name, printTagName(tag), "")
	}

	return name
}

// printProperty converts an object's property name to a valid Go identifier.
func printProperty(p string) string {
	c := strcase.ToCamel(p)
	if c == "Id" {
		c = "ID"
	} else if c == "Ncpus" {
		c = "NCPUs"
	} else if c == "IpAddress" {
		c = "IPAddress"
	} else if c == "UserId" {
		c = "UserID"
	} else if strings.Contains(c, "IdSortMode") {
		strings.ReplaceAll(c, "IdSortMode", "IDSortMode")
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

// cleanPath returns the path as a function we can use for a go template.
func cleanPath(path string) string {
	path = strings.Replace(path, "{", "{{.", -1)
	return strings.Replace(path, "}", "}}", -1)
}

// writeSchemaType writes a type definition for the given schema.
// The additional parameter is only used as a suffix for the type name.
// This is mostly for oneOf types.
func writeSchemaType(f *os.File, name string, s *openapi3.Schema, additionalName string) {
	otype := s.Type
	logrus.Debugf("writing type for schema %q -> %s", name, otype)

	name = printProperty(name)
	typeName := strings.ReplaceAll(strings.TrimSpace(fmt.Sprintf("%s%s", name, printProperty(additionalName))), "Api", "API")

	if len(s.Enum) == 0 && s.OneOf == nil {
		// Write the type description.
		writeSchemaTypeDescription(typeName, s, f)
	}

	if otype == "string" {
		// If this is an enum, write the enum type.
		if len(s.Enum) > 0 {
			// Make sure we don't redeclare the enum type.
			if _, ok := enumStringTypes[makeSingular(typeName)]; !ok {
				// Write the type description.
				writeSchemaTypeDescription(makeSingular(typeName), s, f)

				// Write the enum type.
				fmt.Fprintf(f, "type %s string\n", makeSingular(typeName))

				enumStringTypes[makeSingular(typeName)] = []string{}
			}

			// Define the enum values.
			fmt.Fprintf(f, "const (\n")
			for _, v := range s.Enum {
				// Most likely, the enum values are strings.
				enum, ok := v.(string)
				if !ok {
					logrus.Warnf("TODO: enum value is not a string for %q -> %#v", name, v)
					continue
				}

				// If the enum is empty we want to reflect that in the naming.
				enumName := enum
				if len(enum) <= 0 {
					enumName = "empty"
				}
				// Write the description of the constant.
				fmt.Fprintf(f, "// %s represents the %s `%q`.\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enumName)), makeSingular(name), enumName)
				fmt.Fprintf(f, "\t%s %s = %q\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enumName)), makeSingular(name), enum)

				// Add the enum type to the list of enum types.
				enumStringTypes[makeSingular(typeName)] = append(enumStringTypes[makeSingular(typeName)], enumName)
			}
			// Close the enum values.
			fmt.Fprintf(f, ")\n")

		} else {
			fmt.Fprintf(f, "type %s string\n", name)
		}
	} else if otype == "integer" {
		fmt.Fprintf(f, "type %s int\n", name)
	} else if otype == "number" {
		fmt.Fprintf(f, "type %s float64\n", name)
	} else if otype == "boolean" {
		fmt.Fprintf(f, "type %s bool\n", name)
	} else if otype == "array" {
		fmt.Fprintf(f, "type %s []%s\n", name, s.Items.Value.Type)
	} else if otype == "object" {
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
					writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
				}

				if isLocalObject(v) {
					writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
				}
			}
		}
	} else {
		if s.OneOf != nil {
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
				writeSchemaType(f, name, v.Value, typeName)
				// Add it to our array.
				oneOfTypes = append(oneOfTypes, oneOfType)
			}

			// Now let's create the global oneOf type.
			// Write the type description.
			writeSchemaTypeDescription(typeName, s, f)
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
	}

	// Add a newline at the end of the type.
	fmt.Fprintln(f, "")
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

// toLowerFirstLetter returns the given string with the first letter converted to lower case.
func toLowerFirstLetter(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// makeSingular returns the given string but singular.
func makeSingular(s string) string {
	if strings.HasSuffix(s, "Status") {
		return s
	}
	return strings.TrimSuffix(s, "s")
}

// makePlural returns the given string but plural.
func makePlural(s string) string {
	singular := makeSingular(s)
	if strings.HasSuffix(singular, "s") {
		return singular + "es"
	}

	return singular + "s"
}

// writeSchemaTypeDescription writes the description of the given type.
func writeSchemaTypeDescription(name string, s *openapi3.Schema, f *os.File) {
	if s.Description != "" {
		fmt.Fprintf(f, "// %s is %s\n", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	} else {
		fmt.Fprintf(f, "// %s is the type definition for a %s.\n", name, name)
	}
}

// writeReponseTypeDescription writes the description of the given type.
func writeResponseTypeDescription(name string, r *openapi3.Response, f *os.File) {
	if r.Description != nil {
		fmt.Fprintf(f, "// %s is the response given when %s\n", name, toLowerFirstLetter(
			strings.ReplaceAll(*r.Description, "\n", "\n// ")))
	} else {
		fmt.Fprintf(f, "// %s is the type definition for a %s response.\n", name, name)
	}
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

// writeResponseType writes a type definition for the given response.
func writeResponseType(f *os.File, name string, r *openapi3.Response) {
	// Write the type definition.
	for k, v := range r.Content {
		logrus.Debugf("writing type for response %q -> %q", name, k)

		name := fmt.Sprintf("Response%s", name)

		// Write the type description.
		writeResponseTypeDescription(name, r, f)

		// Print the type definition.
		s := v.Schema
		if s.Ref != "" {
			fmt.Fprintf(f, "type %s %s\n", name, getReferenceSchema(s))
			continue
		}

		writeSchemaType(f, name, s.Value, "")
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func trimStringFromSpace(s string) string {
	if idx := strings.Index(s, " "); idx != -1 {
		return s[:idx]
	}
	return s
}

func containsMatchFirstWord(s []string, str string) bool {
	for _, v := range s {
		if trimStringFromSpace(v) == trimStringFromSpace(str) {
			return true
		}
	}

	return false
}

func isPageParam(s string) bool {
	return s == "nextPage" || s == "pageToken" || s == "limit"
}
