package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/jsondiff"
)

func main() {
	// Load the open API spec from the file.
	wd, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("error getting current working directory: %v", err)
	}
	p := filepath.Join(wd, "spec.json")

	doc, err := openapi3.NewLoader().LoadFromFile(p)
	if err != nil {
		logrus.Fatalf("error loading openAPI spec: %v", err)
	}

	data := Data{
		PackageName:      "kittycad",
		BaseURL:          "https://api.kittycad.io",
		EnvVariable:      "KITTYCAD_API_TOKEN",
		Tags:             []Tag{},
		WorkingDirectory: wd,
		Examples:         []string{},
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
		logrus.Fatalf("error processing template: %v", err)
	}
	data.Examples = append(data.Examples, clientInfo)
	doc.Info.Extensions["x-go"] = map[string]string{
		"install": "go get github.com/kittycad/kittycad.go",
		"client":  clientInfo,
	}

	// Generate the client.go file.
	logrus.Info("Generating client...")
	generateClient(doc, data)

	// Generate the types.go file.
	logrus.Info("Generating types...")
	generateTypes(doc)

	// Generate the responses.go file.
	logrus.Info("Generating responses...")
	generateResponses(doc)

	// Generate the paths.go file.
	logrus.Info("Generating paths...")
	generatePaths(doc)

	// Generate the examples.go file.
	logrus.Info("Generating examples...")
	generateExamplesFile(doc, data)

	// Get the old doc again.
	oldDoc, err := openapi3.NewLoader().LoadFromFile(p)
	if err != nil {
		logrus.Fatalf("error loading openAPI spec: %v", err)
	}
	patch, err := jsondiff.Compare(oldDoc, doc)
	if err != nil {
		logrus.Errorf("error comparing old and new openAPI spec: %v", err)
	}
	patchJson, err := json.MarshalIndent(patch, "", " ")
	if err != nil {
		logrus.Fatalf("error marshalling openAPI spec: %v", err)
	}

	diffFile := filepath.Join(wd, "kittycad.go.patch.json")
	if err := ioutil.WriteFile(diffFile, patchJson, 0644); err != nil {
		logrus.Fatalf("error writing openAPI spec patch to %s: %v", diffFile, err)
	}
}

var EnumStringTypes map[string][]string = map[string][]string{}

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
	for k := range EnumStringTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		enums := EnumStringTypes[name]
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
func generateClient(doc *openapi3.T, data Data) {
	// Generate the lib template.
	if err := processTemplate("lib.tmpl", "lib.go", data); err != nil {
		logrus.Fatalf("error processing template: %v", err)
	}

	// Generate the client template.
	if err := processTemplate("client.tmpl", "client.go", data); err != nil {
		logrus.Fatalf("error processing template: %v", err)
	}
}

func generateExamplesFile(doc *openapi3.T, data Data) {
	// Generate the example template.
	if err := processTemplate("examples.tmpl", "examples.go", data); err != nil {
		logrus.Fatalf("error processing template: %v", err)
	}
}

func printTagName(tag string) string {
	return strings.ReplaceAll(strcase.ToCamel(makeSingular(tag)), "Api", "API")
}

// Generate the paths.go file.
func generatePaths(doc *openapi3.T) {
	f := openGeneratedFile("paths.go")
	defer f.Close()

	// Iterate over all the paths in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, path := range keys {
		p := doc.Paths[path]
		if p.Ref != "" {
			logrus.Warnf("TODO: skipping path for %q, since it is a reference", path)
			continue
		}

		// Ignore the oauth2 paths.
		if strings.HasPrefix(path, "/oauth2/") {
			continue
		}

		writePath(doc, f, path, p)
	}
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

// writePath writes the given path as an http request to the given file.
func writePath(doc *openapi3.T, f *os.File, path string, p *openapi3.PathItem) {
	if p.Get != nil {
		writeMethod(doc, f, http.MethodGet, path, p.Get, false)
	}

	if p.Post != nil {
		writeMethod(doc, f, http.MethodPost, path, p.Post, false)
	}

	if p.Put != nil {
		writeMethod(doc, f, http.MethodPut, path, p.Put, false)
	}

	if p.Delete != nil {
		writeMethod(doc, f, http.MethodDelete, path, p.Delete, false)
	}

	if p.Patch != nil {
		writeMethod(doc, f, http.MethodPatch, path, p.Patch, false)
	}

	if p.Head != nil {
		writeMethod(doc, f, http.MethodHead, path, p.Head, false)
	}
}

func writeMethod(doc *openapi3.T, f *os.File, method string, path string, o *openapi3.Operation, isGetAllPages bool) {
	respType, pagedRespType := getSuccessResponseType(o, isGetAllPages)

	if len(o.Tags) == 0 {
		logrus.Warnf("TODO: skipping operation %q, since it has no tag", o.OperationID)
		return
	}
	tag := printTagName(o.Tags[0])

	fnName := cleanFnName(o.OperationID, tag, path)

	pageResult := false

	// Parse the parameters.
	params := map[string]*openapi3.Parameter{}
	paramsString := ""
	docParamsString := ""
	for index, p := range o.Parameters {
		if p.Ref != "" {
			logrus.Warnf("TODO: skipping parameter for %q, since it is a reference", p.Value.Name)
			continue
		}

		paramName := printPropertyLower(p.Value.Name)

		// Check if we have a page result.
		if isPageParam(paramName) && method == http.MethodGet {
			pageResult = true
		}

		params[p.Value.Name] = p.Value
		paramsString += fmt.Sprintf("%s %s, ", paramName, printType(p.Value.Name, p.Value.Schema))
		if index == len(o.Parameters)-1 {
			docParamsString += fmt.Sprintf("%s", paramName)
		} else {
			docParamsString += fmt.Sprintf("%s, ", paramName)
		}
	}

	if pageResult && isGetAllPages && len(pagedRespType) > 0 {
		respType = pagedRespType
	}

	// Parse the request body.
	reqBodyParam := "nil"
	reqBodyDescription := ""
	if o.RequestBody != nil {
		rb := o.RequestBody

		if rb.Value.Description != "" {
			reqBodyDescription = rb.Value.Description
		}

		if rb.Ref != "" {
			logrus.Warnf("TODO: skipping request body for %q, since it is a reference: %q", path, rb.Ref)
		}

		for mt, r := range rb.Value.Content {
			if mt != "application/json" {
				paramsString += "b io.Reader"
				reqBodyParam = "b"

				if len(docParamsString) > 0 {
					docParamsString += ", "
				}
				docParamsString += "body"
				break
			}

			typeName := printType("", r.Schema)

			paramsString += "j *" + typeName

			if len(docParamsString) > 0 {
				docParamsString += ", "
			}
			docParamsString += "body"

			reqBodyParam = "j"
			break
		}

	}

	ogFnName := fnName
	ogDocParamsString := docParamsString
	if len(pagedRespType) > 0 {
		fnName += "AllPages"
		docParamsString = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(docParamsString, "pageToken", ""), "limit", ""), ", ,", ""))
		paramsString = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(paramsString, "pageToken string", ""), "limit int", ""), ", ,", ""))
		delete(params, "page_token")
		delete(params, "limit")
	}

	logrus.Debugf("writing method %q for path %q -> %q", method, path, fnName)

	var description bytes.Buffer
	// Write the description for the method.
	if o.Summary != "" {
		fmt.Fprintf(&description, "// %s: %s\n", fnName, o.Summary)
	} else {
		fmt.Fprintf(&description, "// %s\n", fnName)
	}
	if o.Description != "" {
		fmt.Fprintln(&description, "//")
		fmt.Fprintf(&description, "// %s\n", strings.ReplaceAll(o.Description, "\n", "\n// "))
	}
	if pageResult && !isGetAllPages {
		fmt.Fprintf(&description, "//\n// To iterate over all pages, use the `%sAllPages` method, instead.\n", fnName)
	}
	if len(pagedRespType) > 0 {
		fmt.Fprintf(&description, "//\n// This method is a wrapper around the `%s` method.\n", ogFnName)
		fmt.Fprintf(&description, "// This method returns all the pages at once.\n")
	}
	if len(params) > 0 {
		fmt.Fprintf(&description, "//\n// Parameters:\n")
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			t := params[name]
			if t.Description != "" {
				fmt.Fprintf(&description, "//\t- `%s`: %s\n", strcase.ToLowerCamel(name), strings.ReplaceAll(t.Description, "\n", "\n//\t\t"))
			} else {
				fmt.Fprintf(&description, "//\t- `%s`\n", strcase.ToLowerCamel(name))
			}
		}
	}

	if reqBodyDescription != "" && reqBodyParam != "nil" {
		fmt.Fprintf(&description, "//\t- `%s`: %s\n", reqBodyParam, strings.ReplaceAll(reqBodyDescription, "\n", "\n// "))
	}

	// Write the description to the file.
	fmt.Fprintf(f, description.String())

	docInfo := map[string]string{
		"example":     fmt.Sprintf("%s", description.String()),
		"libDocsLink": fmt.Sprintf("https://pkg.go.dev/github.com/kittycad/kittycad.go/#%sService.%s", tag, fnName),
	}
	if isGetAllPages {
		og := doc.Paths[path].Get.Extensions["x-go"].(map[string]string)
		docInfo["example"] = fmt.Sprintf("%s\n\n// - OR -\n\n%s", og["example"], docInfo["example"])
		docInfo["libDocsLink"] = fmt.Sprintf("https://pkg.go.dev/github.com/kittycad/kittycad.go/#%sService.%s", tag, ogFnName)
	}

	// Write the method.
	if respType != "" {
		fmt.Fprintf(f, "func (s *%sService) %s(%s) (*%s, error) {\n",
			tag,
			fnName,
			paramsString,
			respType)
		docInfo["example"] += fmt.Sprintf("%s, err := client.%s.%s(%s)", strcase.ToLowerCamel(respType), tag, fnName, docParamsString)
	} else {
		fmt.Fprintf(f, "func (s *%sService) %s(%s) (error) {\n",
			tag,
			fnName,
			paramsString)
		docInfo["example"] += fmt.Sprintf(`if err := client.%s.%s(%s); err != nil {
	panic(err)
}`, tag, fnName, docParamsString)
	}

	// Special case for functions with Base64 helpers.
	if fnName == "CreateConversion" || fnName == "GetConversion" {
		docInfo["example"] = fmt.Sprintf(`%s

// - OR -

// %sWithBase64Helper will automatically base64 encode and decode the contents
// of the file body.
//
// This function is a wrapper around the %s function.
%s, err := client.%s.%sWithBase64Helper(%s)`, docInfo["example"], fnName, fnName, strcase.ToLowerCamel(respType), tag, fnName, docParamsString)
	}

	if method == http.MethodGet {
		doc.Paths[path].Get.Extensions["x-go"] = docInfo
	} else if method == http.MethodPost {
		doc.Paths[path].Post.Extensions["x-go"] = docInfo
	} else if method == http.MethodPut {
		doc.Paths[path].Put.Extensions["x-go"] = docInfo
	} else if method == http.MethodDelete {
		doc.Paths[path].Delete.Extensions["x-go"] = docInfo
	} else if method == http.MethodPatch {
		doc.Paths[path].Patch.Extensions["x-go"] = docInfo
	}

	if len(pagedRespType) > 0 {
		// We want to just recursively call the method for each page.
		fmt.Fprintf(f, `
			var allPages %s
			pageToken := ""
			limit := 100
			for {
				page, err := s.%s(%s)
				if err != nil {
					return nil, err
				}
				allPages = append(allPages, page.Items...)
				if  page.NextPage == "" {
					break
				}
				pageToken = page.NextPage
			}

			return &allPages, nil
		}`, pagedRespType, ogFnName, ogDocParamsString)

		// Return early.
		return
	}

	// Create the url.
	fmt.Fprintln(f, "// Create the url.")
	fmt.Fprintf(f, "path := %q\n", cleanPath(path))
	fmt.Fprintln(f, "uri := resolveRelative(s.client.server, path)")

	if o.RequestBody != nil {
		for mt := range o.RequestBody.Value.Content {
			if mt != "application/json" {
				break
			}

			// We need to encode the request body as json.
			fmt.Fprintln(f, "// Encode the request body as json.")
			fmt.Fprintln(f, "b := new(bytes.Buffer)")
			fmt.Fprintln(f, "if err := json.NewEncoder(b).Encode(j); err != nil {")
			if respType != "" {
				fmt.Fprintln(f, `return nil, fmt.Errorf("encoding json body request failed: %v", err)`)
			} else {
				fmt.Fprintln(f, `return fmt.Errorf("encoding json body request failed: %v", err)`)
			}
			fmt.Fprintln(f, "}")
			reqBodyParam = "b"
			break
		}

	}

	// Create the request.
	fmt.Fprintln(f, "// Create the request.")

	fmt.Fprintf(f, "req, err := http.NewRequest(%q, uri, %s)\n", method, reqBodyParam)
	fmt.Fprintln(f, "if err != nil {")
	if respType != "" {
		fmt.Fprintln(f, `return nil, fmt.Errorf("error creating request: %v", err)`)
	} else {
		fmt.Fprintln(f, `return fmt.Errorf("error creating request: %v", err)`)
	}
	fmt.Fprintln(f, "}")

	// Add the parameters to the url.
	if len(params) > 0 {
		fmt.Fprintln(f, "// Add the parameters to the url.")
		fmt.Fprintln(f, "if err := expandURL(req.URL, map[string]string{")
		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]
			t := printType(name, p.Schema)
			n := printPropertyLower(name)
			if t == "string" {
				fmt.Fprintf(f, "	%q: %s,\n", name, n)
			} else if t == "int" {
				fmt.Fprintf(f, "	%q: strconv.Itoa(%s),\n", name, n)
			} else if t == "float64" {
				fmt.Fprintf(f, "	%q: fmt.Sprintf(\"%%f\", %s),\n", name, n)
			} else {
				fmt.Fprintf(f, "	%q: string(%s),\n", name, n)
			}
		}
		fmt.Fprintln(f, "}); err != nil {")
		if respType != "" {
			fmt.Fprintln(f, `return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)`)
		} else {
			fmt.Fprintln(f, `return fmt.Errorf("expanding URL with parameters failed: %v", err)`)
		}
		fmt.Fprintln(f, "}")
	}

	// Send the request.
	fmt.Fprintln(f, "// Send the request.")
	fmt.Fprintln(f, "resp, err := s.client.client.Do(req)")
	fmt.Fprintln(f, "if err != nil {")
	if respType != "" {
		fmt.Fprintln(f, `return nil, fmt.Errorf("error sending request: %v", err)`)
	} else {
		fmt.Fprintln(f, `return fmt.Errorf("error sending request: %v", err)`)
	}
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f, "defer resp.Body.Close()")

	// Check the response if there were any errors.
	fmt.Fprintln(f, "// Check the response.")
	fmt.Fprintln(f, "if err := checkResponse(resp); err != nil {")
	if respType != "" {
		fmt.Fprintln(f, "return nil, err")
	} else {
		fmt.Fprintln(f, "return err")
	}
	fmt.Fprintln(f, "}")

	if respType != "" {
		// Decode the body from the response.
		fmt.Fprintln(f, "// Decode the body from the response.")
		fmt.Fprintln(f, "if resp.Body == nil {")
		fmt.Fprintln(f, `return nil, errors.New("request returned an empty body in the response")`)
		fmt.Fprintln(f, "}")

		fmt.Fprintf(f, "var body %s\n", respType)
		fmt.Fprintln(f, "if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {")
		fmt.Fprintln(f, `return nil, fmt.Errorf("error decoding response body: %v", err)`)
		fmt.Fprintln(f, "}")

		// Return the response.
		fmt.Fprintln(f, "// Return the response.")
		fmt.Fprintln(f, "return &body, nil")
	} else {
		fmt.Fprintln(f, "// Return.")
		fmt.Fprintln(f, "return nil")
	}

	// Close the method.
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f, "")

	if pageResult && !isGetAllPages {
		// Run the method again with get all pages.
		// Skip doing all pages for now.
		writeMethod(doc, f, method, path, o, true)
	}
}

// cleanPath returns the path as a function we can use for a go template.
func cleanPath(path string) string {
	path = strings.Replace(path, "{", "{{.", -1)
	return strings.Replace(path, "}", "}}", -1)
}

func getSuccessResponseType(o *openapi3.Operation, isGetAllPages bool) (string, string) {
	for name, response := range o.Responses {
		if name == "default" {
			name = "200"
		}

		statusCode, err := strconv.Atoi(strings.ReplaceAll(name, "XX", "00"))
		if err != nil {
			fmt.Printf("error converting %q to an integer: %v\n", name, err)
			os.Exit(1)
		}

		if statusCode < 200 || statusCode >= 300 {
			// Continue early, we just want the successful response.
			continue
		}

		if response.Ref != "" {
			logrus.Warnf("TODO: skipping response for %q, since it is a reference: %q", name, response.Ref)
			continue
		}

		for _, content := range response.Value.Content {
			getAllPagesType := ""
			if isGetAllPages {

				if items, ok := content.Schema.Value.Properties["items"]; ok {
					getAllPagesType = printType("", items)
				} else {
					logrus.Warnf("TODO: skipping response for %q, since it is a get all pages response and has no `items` property:\n%#v", o.OperationID, content.Schema.Value.Properties)
				}
			}
			if content.Schema.Ref != "" {
				return getReferenceSchema(content.Schema), getAllPagesType
			}

			if content.Schema.Value.Title == "Null" {
				return "", ""
			}

			if content.Schema.Value.Type == "array" {
				return printType("", content.Schema), getAllPagesType
			}

			return fmt.Sprintf("Response%s", strcase.ToCamel(o.OperationID)), getAllPagesType
		}
	}

	return "", ""
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
			if _, ok := EnumStringTypes[makeSingular(typeName)]; !ok {
				// Write the type description.
				writeSchemaTypeDescription(makeSingular(typeName), s, f)

				// Write the enum type.
				fmt.Fprintf(f, "type %s string\n", makeSingular(typeName))

				EnumStringTypes[makeSingular(typeName)] = []string{}
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
				EnumStringTypes[makeSingular(typeName)] = append(EnumStringTypes[makeSingular(typeName)], enumName)
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