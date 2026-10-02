package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
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
