package main

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGenerateUUIDArrayQuery(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromData([]byte(`{
		"openapi":"3.0.3","info":{"title":"Array query","version":"1"},"paths":{},
		"components":{"schemas":{"Error":{"type":"object","required":["message"],"properties":{"message":{"type":"string"}}}}}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	var operation openapi3.Operation
	if err := operation.UnmarshalJSON([]byte(`{
		"operationId":"download","tags":["files"],"parameters":[{
			"name":"file_ids","in":"query","explode":false,
			"schema":{"type":"array","items":{"type":"string","format":"uuid"},"minItems":1}
		}],"responses":{"200":{"description":"download","content":{"application/json":{"schema":{"type":"string"}}}}}
	}`)); err != nil {
		t.Fatal(err)
	}
	data := Data{PackageName: "kittycad", Tags: []Tag{{Name: "File"}}, Types: map[string]string{}}
	if err := data.generateMethod(spec, http.MethodGet, "/files/download", &operation, false, spec); err != nil {
		t.Fatal(err)
	}
	if err := data.generateSchemaType("Error", spec.Components.Schemas["Error"].Value, spec); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for filename, template := range map[string]string{"client.go": "client.tmpl", "lib.go": "lib.tmpl", "types.go": "types.tmpl"} {
		source, err := templateToString(template, data)
		if err != nil {
			t.Fatal(err)
		}
		files[filename] = source
	}
	files["paths.go"] = "package kittycad\nimport (\"encoding/json\";\"errors\";\"fmt\";\"net/http\")\n" + strings.Join(data.Paths, "\n")
	for _, filename := range []string{"json_uuid.go", "utils.go"} {
		source, err := goCodeFiles.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		files[filename] = strings.Replace(string(source), "package main", "package kittycad", 1)
	}
	files["paths_test.go"] = `package kittycad
import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)
func TestDownloadSelection(t *testing.T) {
	const first = "f8e50cf2-a25a-4e88-baaa-96cd5e551242"
	const second = "9e84c4a5-29d3-48dc-9c59-c3a44e7ccbf0"
	for _, test := range []struct {
		name string
		ids []UUID
		present bool
		value string
	}{
		{"omitted", nil, false, ""},
		{"empty", []UUID{}, true, ""},
		{"one", []UUID{ParseUUID(first)}, true, first},
		{"two", []UUID{ParseUUID(first), ParseUUID(second)}, true, first + "," + second},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/files/download" || r.URL.Query().Has("file_ids") != test.present || r.URL.Query().Get("file_ids") != test.value {
					t.Errorf("unexpected download selection: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if test.present && test.value == "" {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("{\"message\":\"empty file selection\"}"))
					return
				}
				_, _ = w.Write([]byte("\"download\""))
			}))
			defer server.Close()
			client, err := NewClient("test-token", "test")
			if err != nil { t.Fatal(err) }
			if err := client.WithBaseURL(server.URL); err != nil { t.Fatal(err) }
			response, err := client.File.Download(test.ids)
			if test.present && test.value == "" {
				var httpError *HTTPError
				if !errors.As(err, &httpError) || httpError.StatusCode != http.StatusBadRequest { t.Fatalf("expected empty-selection HTTP400, got %v", err) }
				return
			}
			if err != nil { t.Fatal(err) }
			if response == nil || *response != "download" { t.Fatalf("unexpected download response: %v", response) }
		})
	}
}
`
	testGeneratedPackage(t, files)
}

func testGeneratedPackage(t *testing.T, files map[string]string) {
	t.Helper()
	for _, filename := range []string{"go.mod", "go.sum"} {
		contents, err := os.ReadFile(filepath.Join("..", filename))
		if err != nil {
			t.Fatal(err)
		}
		files[filename] = string(contents)
	}
	directory := t.TempDir()
	for filename, contents := range files {
		if err := os.WriteFile(filepath.Join(directory, filename), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("go", "test", "-mod=readonly", "./...")
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generated package failed to compile or satisfy its consumer contract: %v\n%s", err, output)
	}
}
