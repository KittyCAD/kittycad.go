package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGenerateInlineObjectWithoutProperties(t *testing.T) {
	for _, test := range []struct {
		name     string
		property string
		wantType string
	}{
		{"open", `{"type":"object"}`, "map[string]any"},
		{"closed", `{"type":"object","additionalProperties":false}`, "struct{}"},
		{"typed map", `{"type":"object","additionalProperties":{"type":"string"}}`, "map[string]string"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var schema openapi3.Schema
			if err := json.Unmarshal([]byte(`{"type":"object","required":["centroid"],"properties":{"centroid":`+test.property+`}}`), &schema); err != nil {
				t.Fatal(err)
			}
			data := Data{PackageName: "generated", Types: map[string]string{}}
			spec := &openapi3.T{Components: &openapi3.Components{Schemas: openapi3.Schemas{}}}
			if err := data.generateSchemaType("Position", &schema, spec); err != nil {
				t.Fatal(err)
			}
			source, err := templateToString("types.tmpl", data)
			if err != nil {
				t.Fatal(err)
			}
			files := token.NewFileSet()
			file, err := parser.ParseFile(files, "types.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			config := types.Config{}
			pkg, err := config.Check("generated", files, []*ast.File{file}, nil)
			if err != nil {
				t.Fatal(err)
			}
			position := pkg.Scope().Lookup("Position").Type().Underlying().(*types.Struct)
			if got := position.Field(0).Type().String(); got != test.wantType {
				t.Fatalf("centroid type = %q, want %q", got, test.wantType)
			}
		})
	}
}

func TestGenerateArrayOfUUIDReferences(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromData([]byte(`{
		"openapi":"3.0.3","info":{"title":"UUID references","version":"1"},"paths":{},
		"components":{"schemas":{
			"ModelingCmdId":{"type":"string","format":"uuid"},
			"Label":{"type":"string"},
			"Projection":{"type":"string","enum":["orthographic"]},
			"Batch":{"type":"object","required":["commands","labels","projections"],"properties":{
				"commands":{"type":"array","items":{"$ref":"#/components/schemas/ModelingCmdId"}},
				"labels":{"type":"array","items":{"$ref":"#/components/schemas/Label"}},
				"projections":{"type":"array","items":{"$ref":"#/components/schemas/Projection"}}
			}}
		}}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	data := Data{PackageName: "kittycad", Types: map[string]string{}}
	for _, name := range []string{"ModelingCmdId", "Label", "Projection", "Batch"} {
		if err := data.generateSchemaType(name, spec.Components.Schemas[name].Value, spec); err != nil {
			t.Fatal(err)
		}
	}
	source, err := templateToString("types.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	uuidSource, err := goCodeFiles.ReadFile("json_uuid.go")
	if err != nil {
		t.Fatal(err)
	}
	example, err := data.generateExampleValue("Batch", spec.Components.Schemas["Batch"], spec, true)
	if err != nil {
		t.Fatal(err)
	}
	consumer := `package kittycad_test
import (
	"encoding/json"
	"testing"
	"github.com/kittycad/kittycad.go"
)
var _ = GENERATED_EXAMPLE
func TestGeneratedUUIDArrayRoundTrip(t *testing.T) {
	const id = "f8e50cf2-a25a-4e88-baaa-96cd5e551242"
	batch := kittycad.Batch{
		Commands: []kittycad.UUID{kittycad.ParseUUID(id)},
		Labels: []string{"example"},
		Projections: []kittycad.Projection{kittycad.ProjectionOrthographic},
	}
	wire, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	if string(wire) != "{\"commands\":[\""+id+"\"],\"labels\":[\"example\"],\"projections\":[\"orthographic\"]}" {
		t.Fatalf("unexpected JSON: %s", wire)
	}
	var decoded kittycad.Batch
	if err := json.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Commands) != 1 || decoded.Commands[0].String() != id || len(decoded.Labels) != 1 || decoded.Labels[0] != "example" || len(decoded.Projections) != 1 || decoded.Projections[0] != kittycad.ProjectionOrthographic {
		t.Fatalf("unexpected decoded batch: %#v", decoded)
	}
}
`
	directory := t.TempDir()
	files := map[string][]byte{
		"types.go":      []byte(source),
		"json_uuid.go":  []byte(strings.Replace(string(uuidSource), "package main", "package kittycad", 1)),
		"types_test.go": []byte(strings.Replace(consumer, "GENERATED_EXAMPLE", example, 1)),
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		contents, err := os.ReadFile(filepath.Join("..", name))
		if err != nil {
			t.Fatal(err)
		}
		files[name] = contents
	}
	for name, contents := range files {
		filename := filepath.Join(directory, name)
		if err := os.WriteFile(filename, contents, 0600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("go", "test", "-mod=readonly", "./...")
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generated UUID arrays failed to compile or round-trip: %v\n%s", err, output)
	}
}
