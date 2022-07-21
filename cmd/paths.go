package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
)

// Generate the paths.go file.
func (data *Data) generatePaths(doc *openapi3.T) error {
	// Iterate over all the paths in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, pathName := range keys {
		path := doc.Paths[pathName]
		if path.Ref != "" {
			logrus.Warnf("TODO: skipping path for %q, since it is a reference", pathName)
			continue
		}

		if err := data.generatePath(doc, pathName, path, doc); err != nil {
			return err
		}
	}

	// Write the paths to the template.
	if err := processTemplate("paths.tmpl", "paths.go", *data); err != nil {
		return err
	}

	return nil
}

// generatePath writes the given path as an http request to the given file.
func (data *Data) generatePath(doc *openapi3.T, pathName string, path *openapi3.PathItem, spec *openapi3.T) error {
	if path.Get != nil {
		if err := data.generateMethod(doc, http.MethodGet, pathName, path.Get, false, spec); err != nil {
			return err
		}
	}

	if path.Post != nil {
		if err := data.generateMethod(doc, http.MethodPost, pathName, path.Post, false, spec); err != nil {
			return err
		}
	}

	if path.Put != nil {
		if err := data.generateMethod(doc, http.MethodPut, pathName, path.Put, false, spec); err != nil {
			return err
		}
	}

	if path.Delete != nil {
		if err := data.generateMethod(doc, http.MethodDelete, pathName, path.Delete, false, spec); err != nil {
			return err
		}
	}

	if path.Patch != nil {
		if err := data.generateMethod(doc, http.MethodPatch, pathName, path.Patch, false, spec); err != nil {
			return err
		}
	}

	if path.Head != nil {
		if err := data.generateMethod(doc, http.MethodHead, pathName, path.Head, false, spec); err != nil {
			return err
		}
	}

	return nil
}

// Path holds what we need for generating our functions.
type Path struct {
	Name        string
	Tag         string
	Method      string
	Path        string
	Description string
	RequestBody *RequestBody
	Args        []Arg
	Response    *Response
	PackageName string
}

func (function Path) getDescription(operation *openapi3.Operation) string {
	// Write the description for the method.
	description := ""
	if operation.Summary != "" {
		description = fmt.Sprintf("%s: %s\n", function.Name, operation.Summary)
	} else {
		description = fmt.Sprintf("%s makes a `%s` request to `%s`.\n", function.Name, function.Method, function.Path)
	}

	if operation.Description != "" {
		description = fmt.Sprintf("%s\n%s\n", description, operation.Description)
	}
	if len(function.Args) > 0 || function.RequestBody != nil {
		description = fmt.Sprintf("%s\n\nParameters\n\n", description)
	}

	for _, arg := range function.Args {
		if arg.Description != "" {
			description = fmt.Sprintf("%s\t- `%s`: %s\n", description, arg.Name, strings.ReplaceAll(arg.Description, "\n", "\n\t\t"))
		} else {
			description = fmt.Sprintf("%s\t- `%s`\n", description, arg.Name)
		}
	}
	if function.RequestBody != nil {
		if function.RequestBody.Description != "" {
			description = fmt.Sprintf("%s\t- `body`: %s\n", description, strings.ReplaceAll(function.RequestBody.Description, "\n", "\n\t\t"))
		} else {
			description = fmt.Sprintf("%s\t- `body`\n", description)
		}
	}

	return strings.ReplaceAll(description, "\n", "\n// ")
}

// Arg is an argument to a path function.
type Arg struct {
	Name        string
	Description string
	Property    string
	Type        string
	ToString    string
	Required    bool
	Example     string
}

// RequestBody is a request body for a path function.
type RequestBody struct {
	Type        string
	Description string
	MediaType   string
	Example     string
}

// Response is a response for a path function.
type Response struct {
	Type string
}

func (data *Data) generateMethod(doc *openapi3.T, method string, pathName string, operation *openapi3.Operation, isGetAllPages bool, spec *openapi3.T) error {
	if len(operation.Tags) == 0 {
		return fmt.Errorf("operation at %q %q has no tags", pathName, method)
	}

	tag := printTagName(operation.Tags[0])
	function := Path{
		Name:        cleanFnName(operation.OperationID, tag, pathName),
		Tag:         printProperty(tag),
		Path:        cleanPath(pathName),
		Method:      method,
		Args:        []Arg{},
		PackageName: data.PackageName,
	}

	logrus.Debugf("writing method %q for path %q -> %q", method, pathName, function.Name)

	// Get the response type for the function.
	respType, _, err := getSuccessResponseType(operation, isGetAllPages, spec)
	if err != nil {
		return err
	}
	if respType != "" {
		function.Response = &Response{Type: respType}
	}

	// Parse the parameters.
	for _, p := range operation.Parameters {
		if p.Ref != "" {
			return fmt.Errorf("parameter for %q %q, is a reference: %q, not yet handled", pathName, method, p.Ref)
		}

		// Get the type for the parameter.
		typeName, err := printType(p.Value.Name, p.Value.Schema, spec)
		if err != nil {
			return err
		}

		description, err := getDescriptionForSchemaOrReference(p.Value.Schema, spec)
		if err != nil {
			return err
		}

		example, err := data.generateExampleValue(p.Value.Name, p.Value.Schema, spec, true)
		if err != nil {
			return err
		}

		// Ready ourselves for adding our arg.
		arg := Arg{
			Name:        printPropertyLower(p.Value.Name),
			Property:    p.Value.Name,
			Description: description,
			Type:        typeName,
			Required:    p.Value.Required,
			Example:     example,
		}

		if typeName == "string" {
			arg.ToString = arg.Name
		} else if typeName == "int" {
			arg.ToString = fmt.Sprintf("strconv.Itoa(%s)", arg.Name)
		} else if typeName == "float64" {
			arg.ToString = fmt.Sprintf("fmt.Sprintf(\"%%f\", %s)", arg.Name)
		} else if isTypeToString(typeName) {
			arg.ToString = fmt.Sprintf("%s.String()", arg.Name)
		} else {
			arg.ToString = fmt.Sprintf("string(%s)", arg.Name)
		}

		// Add our arg to the function.
		function.Args = append(function.Args, arg)
	}

	// Parse the request body.
	if operation.RequestBody != nil {
		if operation.RequestBody.Ref != "" {
			return fmt.Errorf("request body for %q %q, is a reference: %q, not yet handled", pathName, method, operation.RequestBody.Ref)
		}

		for mt, r := range operation.RequestBody.Value.Content {
			typeName, err := printType("", r.Schema, spec)
			if err != nil {
				return err
			}

			example, err := data.generateExampleValue(typeName, r.Schema, spec, true)
			if err != nil {
				return err
			}

			// Add our request body to the function.
			function.RequestBody = &RequestBody{
				Type:      typeName,
				MediaType: mt,
				Example:   example,
			}

			description, err := getDescriptionForSchemaOrReference(r.Schema, spec)
			if err != nil {
				return err
			}

			if description != "" {
				function.RequestBody.Description = description
			}

			break
		}
	}

	// Now we can get the description since we have filled in everything else.
	function.Description = function.getDescription(operation)

	// Build the example function.
	example, err := templateToString("function-example.tmpl", function)
	if err != nil {
		return err
	}
	data.Examples = append(data.Examples, example)

	// Print the template for the function.
	f, err := templateToString("path.tmpl", function)
	if err != nil {
		return err
	}

	// Add the function to our list of functions.
	data.Paths = append(data.Paths, f)

	// Add it to our docs.
	docInfo := map[string]string{
		"example":     fmt.Sprintf("// %s\n%s", function.getDescription(operation), example),
		"libDocsLink": fmt.Sprintf("https://pkg.go.dev/github.com/kittycad/kittycad.go/#%sService.%s", function.Tag, function.Name),
	}
	if method == http.MethodGet {
		doc.Paths[pathName].Get.Extensions["x-go"] = docInfo
	} else if method == http.MethodPost {
		doc.Paths[pathName].Post.Extensions["x-go"] = docInfo
	} else if method == http.MethodPut {
		doc.Paths[pathName].Put.Extensions["x-go"] = docInfo
	} else if method == http.MethodDelete {
		doc.Paths[pathName].Delete.Extensions["x-go"] = docInfo
	} else if method == http.MethodPatch {
		doc.Paths[pathName].Patch.Extensions["x-go"] = docInfo
	}

	return nil
}

func getSuccessResponseType(o *openapi3.Operation, isGetAllPages bool, spec *openapi3.T) (string, string, error) {
	for name, response := range o.Responses {
		if name == "default" {
			name = "200"
		}

		statusCode, err := strconv.Atoi(strings.ReplaceAll(name, "XX", "00"))
		if err != nil {
			return "", "", fmt.Errorf("converting %q to an integer failed: %v", name, err)
		}

		if statusCode < 200 || statusCode >= 300 {
			// Continue early, we just want the successful response.
			continue
		}

		if response.Ref != "" {
			return "", "", fmt.Errorf("response for %q, is a reference: %q", name, response.Ref)
		}

		for _, content := range response.Value.Content {
			getAllPagesType := ""
			if isGetAllPages {
				if items, ok := content.Schema.Value.Properties["items"]; ok {
					getAllPagesType, err = printType("", items, spec)
					if err != nil {
						return "", "", err
					}
				} else {
					return "", "", fmt.Errorf("TODO: skipping response for %q, since it is a get all pages response and has no `items` property:\n%#v", o.OperationID, content.Schema.Value.Properties)
				}
			}
			if content.Schema.Ref != "" {
				return getReferenceSchema(content.Schema), getAllPagesType, nil
			}

			if content.Schema.Value.Title == "Null" {
				return "", "", nil
			}

			if content.Schema.Value.Type == "array" {
				t, err := printType("", content.Schema, spec)
				if err != nil {
					return "", "", err
				}
				return t, getAllPagesType, nil
			}

			// Get the type for the schema.
			t, err := printType("", content.Schema, spec)
			if err != nil {
				return "", "", err
			}

			// If it's an interface then it was an empty schema and therefore there is no response.
			if t == "any" {
				return "", getAllPagesType, nil
			}

			return t, getAllPagesType, nil
		}
	}

	// This endpoint does not have a response.
	return "", "", nil
}

func cleanFnName(name string, tag string, path string) string {
	name = printProperty(name)

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

	if strings.Contains(name, printTagName(tag)) {
		name = strings.ReplaceAll(name, printTagName(tag)+"s", "")
		name = strings.ReplaceAll(name, printTagName(tag), "")
	}

	return name
}

// cleanPath returns the path as a function we can use for a go template.
func cleanPath(path string) string {
	path = strings.Replace(path, "{", "{{.", -1)
	return strings.Replace(path, "}", "}}", -1)
}

func getDescriptionForSchemaOrReference(ref *openapi3.SchemaRef, spec *openapi3.T) (string, error) {
	if ref.Ref != "" {
		// Get the schema for the reference.
		r := strings.TrimPrefix(ref.Ref, "#/components/schemas/")
		reference, ok := spec.Components.Schemas[r]
		if !ok {
			return "", fmt.Errorf("reference %q not found in schemas", r)
		}
		return getDescriptionForSchemaOrReference(reference, spec)
	}

	return ref.Value.Description, nil
}
