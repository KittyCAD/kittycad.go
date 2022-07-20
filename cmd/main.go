package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
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
