package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/jsondiff"
)

// Embed our go code files.
//
//go:embed *.go
var goCodeFiles embed.FS

func init() {
	// results in "Api" using ToCamel("API")
	// results in "api" using ToLowerCamel("API")
	strcase.ConfigureAcronym("API", "api")
	strcase.ConfigureAcronym("UUID", "uuid")
	strcase.ConfigureAcronym("URL", "url")
	strcase.ConfigureAcronym("IP", "ip")
	strcase.ConfigureAcronym("ID", "id")
	strcase.ConfigureAcronym("JSON", "json")
	strcase.ConfigureAcronym("GPU", "gpu")

}

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
		Examples:         []string{},
		Paths:            []string{},
		Types:            map[string]string{},
		WorkingDirectory: wd,
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
	if err := data.generateTypes(doc); err != nil {
		return err
	}

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

	// Generate our files that are the same as our source files.
	if err := generateSourceFiles(data); err != nil {
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

	// Make sure we are not doing any "Remove" operations.
	newPatch := jsondiff.Patch{}
	for i, d := range patch {
		if d.Type != jsondiff.OperationRemove {
			newPatch = append(newPatch, patch[i])
		}
	}

	patchJSON, err := json.MarshalIndent(newPatch, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling openAPI spec: %v", err)
	}

	diffFile := filepath.Join(wd, "kittycad.go.patch.json")
	if err := os.WriteFile(diffFile, patchJSON, 0644); err != nil {
		return fmt.Errorf("error writing openAPI spec patch to %s: %v", diffFile, err)
	}

	return nil
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

func generateSourceFiles(data Data) error {
	sourceFiles := []string{
		"json_base64.go",
		"json_base64_test.go",
		"json_ip.go",
		"json_ip_test.go",
		"json_time.go",
		"json_time_test.go",
		"json_url.go",
		"json_url_test.go",
		"json_uuid.go",
		"json_uuid_test.go",
		"utils.go",
		"utils_test.go",
	}

	for _, sourceFile := range sourceFiles {
		// Read the file.
		contents, err := goCodeFiles.ReadFile(sourceFile)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", sourceFile, err)
		}
		contentsString := strings.Replace(string(contents), "package main", fmt.Sprintf("package %s", data.PackageName), 1)
		// Write the source file to the current directory.
		if err := os.WriteFile(filepath.Join(data.WorkingDirectory, sourceFile), []byte(contentsString), 0644); err != nil {
			return fmt.Errorf("error writing file %s: %v", sourceFile, err)
		}
	}

	return nil
}
