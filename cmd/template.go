package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"html/template"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Embed the entire directory.
//go:embed tmpl
var templateFiles embed.FS

// Data holds information for templates.
type Data struct {
	PackageName      string
	BaseURL          string
	EnvVariable      string
	Tags             []Tag
	WorkingDirectory string
	Examples         []string
}

// Tag holds information about tags.
type Tag struct {
	Name        string
	Description string
}

func templateToString(templateName string, data Data) (string, error) {
	tmpl := template.Must(template.New("").ParseFS(templateFiles, path.Join("tmpl", templateName)))
	var processed bytes.Buffer
	err := tmpl.ExecuteTemplate(&processed, templateName, data)
	if err != nil {
		return "", fmt.Errorf("error executing template %q: %v", templateName, err)
	}
	formatted, err := format.Source(processed.Bytes())
	if err != nil {
		return "", fmt.Errorf("error formatting template %q output: %v", templateName, err)
	}

	return string(formatted), nil
}

func processTemplate(templateName string, outputFile string, data Data) error {
	formatted, err := templateToString(templateName, data)
	if err != nil {
		return err
	}

	fmt.Println(formatted)

	outputPath := filepath.Join(data.WorkingDirectory, outputFile)
	logrus.Debugf("Writing file: ", outputPath)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file %q: %v", outputPath, err)
	}

	w := bufio.NewWriter(f)
	w.WriteString(formatted)

	w.Flush()

	return nil
}
