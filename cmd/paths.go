package main

import (
	"bytes"
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

func (data *Data) generateMethod(doc *openapi3.T, method string, pathName string, operation *openapi3.Operation, isGetAllPages bool, spec *openapi3.T) error {
	respType, pagedRespType, err := getSuccessResponseType(operation, isGetAllPages, spec)
	if err != nil {
		return err
	}

	if len(operation.Tags) == 0 {
		return fmt.Errorf("operation at %q %q has no tags", pathName, method)
	}
	tag := printTagName(operation.Tags[0])

	fnName := cleanFnName(operation.OperationID, tag, pathName)

	pageResult := false

	// Parse the parameters.
	params := map[string]*openapi3.Parameter{}
	paramsString := ""
	docParamsString := ""
	for index, p := range operation.Parameters {
		if p.Ref != "" {
			return fmt.Errorf("parameter for %q %q, is a reference: %q, not yet handled", pathName, method, p.Ref)
		}

		paramName := printPropertyLower(p.Value.Name)

		// Check if we have a page result.
		if isPageParam(paramName) && method == http.MethodGet {
			pageResult = true
		}

		typeName, err := printType(p.Value.Name, p.Value.Schema, spec)
		if err != nil {
			return err
		}

		params[p.Value.Name] = p.Value
		paramsString += fmt.Sprintf("%s %s, ", paramName, typeName)
		if index == len(operation.Parameters)-1 {
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
	if operation.RequestBody != nil {
		rb := operation.RequestBody

		if rb.Value.Description != "" {
			reqBodyDescription = rb.Value.Description
		}

		if rb.Ref != "" {
			return fmt.Errorf("request body for %q %q, is a reference: %q, not yet handled", pathName, method, rb.Ref)
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

			typeName, err := printType("", r.Schema, spec)
			if err != nil {
				return err
			}

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

	logrus.Debugf("writing method %q for path %q -> %q", method, pathName, fnName)

	var description bytes.Buffer
	// Write the description for the method.
	if operation.Summary != "" {
		fmt.Fprintf(&description, "// %s: %s\n", fnName, operation.Summary)
	} else {
		fmt.Fprintf(&description, "// %s\n", fnName)
	}
	if operation.Description != "" {
		fmt.Fprintln(&description, "//")
		fmt.Fprintf(&description, "// %s\n", strings.ReplaceAll(operation.Description, "\n", "\n// "))
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

	var f bytes.Buffer

	// Write the description to the file.
	fmt.Fprintf(&f, description.String())

	docInfo := map[string]string{
		"example":     fmt.Sprintf("%s", description.String()),
		"libDocsLink": fmt.Sprintf("https://pkg.go.dev/github.com/kittycad/kittycad.go/#%sService.%s", tag, fnName),
	}
	if isGetAllPages {
		og := doc.Paths[pathName].Get.Extensions["x-go"].(map[string]string)
		docInfo["example"] = fmt.Sprintf("%s\n\n// - OR -\n\n%s", og["example"], docInfo["example"])
		docInfo["libDocsLink"] = fmt.Sprintf("https://pkg.go.dev/github.com/kittycad/kittycad.go/#%sService.%s", tag, ogFnName)
	}

	// Write the method.
	if respType != "" {
		fmt.Fprintf(&f, "func (s *%sService) %s(%s) (*%s, error) {\n",
			tag,
			fnName,
			paramsString,
			respType)
		docInfo["example"] += fmt.Sprintf("%s, err := client.%s.%s(%s)", strcase.ToLowerCamel(respType), tag, fnName, docParamsString)
	} else {
		fmt.Fprintf(&f, "func (s *%sService) %s(%s) (error) {\n",
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

	if len(pagedRespType) > 0 {
		// We want to just recursively call the method for each page.
		fmt.Fprintf(&f, `
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
		return nil
	}

	// Create the url.
	fmt.Fprintln(&f, "// Create the url.")
	fmt.Fprintf(&f, "path := %q\n", cleanPath(pathName))
	fmt.Fprintln(&f, "uri := resolveRelative(s.client.server, path)")

	if operation.RequestBody != nil {
		for mt := range operation.RequestBody.Value.Content {
			if mt != "application/json" {
				break
			}

			// We need to encode the request body as json.
			fmt.Fprintln(&f, "// Encode the request body as json.")
			fmt.Fprintln(&f, "b := new(bytes.Buffer)")
			fmt.Fprintln(&f, "if err := json.NewEncoder(b).Encode(j); err != nil {")
			if respType != "" {
				fmt.Fprintln(&f, `return nil, fmt.Errorf("encoding json body request failed: %v", err)`)
			} else {
				fmt.Fprintln(&f, `return fmt.Errorf("encoding json body request failed: %v", err)`)
			}
			fmt.Fprintln(&f, "}")
			reqBodyParam = "b"
			break
		}

	}

	// Create the request.
	fmt.Fprintln(&f, "// Create the request.")

	fmt.Fprintf(&f, "req, err := http.NewRequest(%q, uri, %s)\n", method, reqBodyParam)
	fmt.Fprintln(&f, "if err != nil {")
	if respType != "" {
		fmt.Fprintln(&f, `return nil, fmt.Errorf("error creating request: %v", err)`)
	} else {
		fmt.Fprintln(&f, `return fmt.Errorf("error creating request: %v", err)`)
	}
	fmt.Fprintln(&f, "}")

	// Add the parameters to the url.
	if len(params) > 0 {
		fmt.Fprintln(&f, "// Add the parameters to the url.")
		fmt.Fprintln(&f, "if err := expandURL(req.URL, map[string]string{")
		// Iterate over all the paths in the spec and write the types.
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			p := params[name]

			t, err := printType(name, p.Schema, spec)
			if err != nil {
				return err
			}

			n := printPropertyLower(name)
			if t == "string" {
				fmt.Fprintf(&f, "	%q: %s,\n", name, n)
			} else if t == "int" {
				fmt.Fprintf(&f, "	%q: strconv.Itoa(%s),\n", name, n)
			} else if t == "float64" {
				fmt.Fprintf(&f, "	%q: fmt.Sprintf(\"%%f\", %s),\n", name, n)
			} else if isTypeToString(t) {
				fmt.Fprintf(&f, "	%q: %s.String(),\n", name, n)
			} else {
				fmt.Fprintf(&f, "	%q: string(%s),\n", name, n)
			}
		}
		fmt.Fprintln(&f, "}); err != nil {")
		if respType != "" {
			fmt.Fprintln(&f, `return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)`)
		} else {
			fmt.Fprintln(&f, `return fmt.Errorf("expanding URL with parameters failed: %v", err)`)
		}
		fmt.Fprintln(&f, "}")
	}

	// Send the request.
	fmt.Fprintln(&f, "// Send the request.")
	fmt.Fprintln(&f, "resp, err := s.client.client.Do(req)")
	fmt.Fprintln(&f, "if err != nil {")
	if respType != "" {
		fmt.Fprintln(&f, `return nil, fmt.Errorf("error sending request: %v", err)`)
	} else {
		fmt.Fprintln(&f, `return fmt.Errorf("error sending request: %v", err)`)
	}
	fmt.Fprintln(&f, "}")
	fmt.Fprintln(&f, "defer resp.Body.Close()")

	// Check the response if there were any errors.
	fmt.Fprintln(&f, "// Check the response.")
	fmt.Fprintln(&f, "if err := checkResponse(resp); err != nil {")
	if respType != "" {
		fmt.Fprintln(&f, "return nil, err")
	} else {
		fmt.Fprintln(&f, "return err")
	}
	fmt.Fprintln(&f, "}")

	if respType != "" {
		// Decode the body from the response.
		fmt.Fprintln(&f, "// Decode the body from the response.")
		fmt.Fprintln(&f, "if resp.Body == nil {")
		fmt.Fprintln(&f, `return nil, errors.New("request returned an empty body in the response")`)
		fmt.Fprintln(&f, "}")

		fmt.Fprintf(&f, "var body %s\n", respType)
		fmt.Fprintln(&f, "if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {")
		fmt.Fprintln(&f, `return nil, fmt.Errorf("error decoding response body: %v", err)`)
		fmt.Fprintln(&f, "}")

		// Return the response.
		fmt.Fprintln(&f, "// Return the response.")
		fmt.Fprintln(&f, "return &body, nil")
	} else {
		fmt.Fprintln(&f, "// Return.")
		fmt.Fprintln(&f, "return nil")
	}

	// Close the method.
	fmt.Fprintln(&f, "}")
	fmt.Fprintln(&f, "")

	// Add the function to our list of functions.
	data.Paths = append(data.Paths, f.String())

	if pageResult && !isGetAllPages {
		// Run the method again with get all pages.
		// Skip doing all pages for now.
		data.generateMethod(doc, method, pathName, operation, true, spec)
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
			logrus.Warnf("TODO: skipping response for %q, since it is a reference: %q", name, response.Ref)
			continue
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
					logrus.Warnf("TODO: skipping response for %q, since it is a get all pages response and has no `items` property:\n%#v", o.OperationID, content.Schema.Value.Properties)
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

			return fmt.Sprintf("Response%s", strcase.ToCamel(o.OperationID)), getAllPagesType, nil
		}
	}

	// This endpoint does not have a response.
	return "", "", nil
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

// cleanPath returns the path as a function we can use for a go template.
func cleanPath(path string) string {
	path = strings.Replace(path, "{", "{{.", -1)
	return strings.Replace(path, "}", "}}", -1)
}

func isPageParam(s string) bool {
	return s == "nextPage" || s == "pageToken" || s == "limit"
}
