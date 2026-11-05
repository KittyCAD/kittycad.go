package main

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestMigrationVariantsPreserveReferencedModels(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromData([]byte(`{
		"openapi":"3.0.3","info":{"title":"Migration messages","version":"1"},"paths":{},
		"components":{"schemas":{
			"KclMigrationApplication":{"type":"object","required":["revision"],"properties":{"revision":{"type":"integer"}}},
			"KclMigrationOperation":{"type":"object","required":["id"],"properties":{"id":{"type":"string"}}},
			"KclMigrationServerMessage":{"oneOf":[
				{"type":"object","required":["type","application"],"properties":{"type":{"type":"string","enum":["application"]},"application":{"$ref":"#/components/schemas/KclMigrationApplication"}}},
				{"type":"object","required":["type","operation"],"properties":{"type":{"type":"string","enum":["operation"]},"operation":{"$ref":"#/components/schemas/KclMigrationOperation"}}}
			]},
			"MlCopilotServerMessage":{"oneOf":[
				{"type":"object","required":["kcl_migration_application"],"properties":{"kcl_migration_application":{"type":"object","required":["application"],"properties":{"application":{"$ref":"#/components/schemas/KclMigrationApplication"}}}}},
				{"type":"object","required":["kcl_migration_operation"],"properties":{"kcl_migration_operation":{"type":"object","required":["operation"],"properties":{"operation":{"$ref":"#/components/schemas/KclMigrationOperation"}}}}}
			]}
		}}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	data := Data{PackageName: "kittycad", Types: map[string]string{}}
	for _, name := range []string{"KclMigrationApplication", "KclMigrationOperation", "KclMigrationServerMessage", "MlCopilotServerMessage"} {
		if err := data.generateSchemaType(name, spec.Components.Schemas[name].Value, spec); err != nil {
			t.Fatal(err)
		}
	}
	source, err := templateToString("types.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	example, err := data.generateExampleValue("MlCopilotServerMessageKclMigrationApplication", spec.Components.Schemas["MlCopilotServerMessage"].Value.OneOf[0], spec, true)
	if err != nil {
		t.Fatal(err)
	}
	testGeneratedPackage(t, map[string]string{
		"types.go": source,
		"types_test.go": strings.Replace(`package kittycad_test
import ("encoding/json"; "testing"; "github.com/kittycad/kittycad.go")
var _ = GENERATED_EXAMPLE
func TestMigrationMessages(t *testing.T) {
	var application kittycad.KclMigrationServerMessageApplication
	if err := json.Unmarshal([]byte("{\"type\":\"application\",\"application\":{\"revision\":3}}"), &application); err != nil { t.Fatal(err) }
	if application.Type != "application" || application.Application.Revision != 3 { t.Fatalf("lost application acknowledgement: %+v", application) }
	var operation kittycad.KclMigrationServerMessageOperation
	if err := json.Unmarshal([]byte("{\"type\":\"operation\",\"operation\":{\"id\":\"migration\"}}"), &operation); err != nil { t.Fatal(err) }
	if operation.Type != "operation" || operation.Operation.ID != "migration" { t.Fatalf("lost migration outcome: %+v", operation) }
	var embedded kittycad.MlCopilotServerMessageKclMigrationApplication
	if err := json.Unmarshal([]byte("{\"kcl_migration_application\":{\"application\":{\"revision\":3}}}"), &embedded); err != nil { t.Fatal(err) }
	if embedded.KclMigrationApplication.Application.Revision != 3 { t.Fatalf("lost embedded acknowledgement: %+v", embedded) }
}
`, "GENERATED_EXAMPLE", example, 1),
	})
}
