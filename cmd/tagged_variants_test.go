package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestCopilotVariantNamesSurvivePayloadGrowth(t *testing.T) {
	for _, extra := range []string{"", `"aaa_new_field":{"type":"string"},`} {
		var schema openapi3.Schema
		input := `{"oneOf":[{"type":"object","required":["type","content"],"properties":{` + extra + `"content":{"type":"string"},"type":{"type":"string","enum":["user"]}}},{"type":"object","required":["type"],"properties":{"type":{"type":"string","enum":["ping"]}}}]}`
		if err := json.Unmarshal([]byte(input), &schema); err != nil {
			t.Fatal(err)
		}
		data := Data{Types: map[string]string{}}
		if err := data.generateOneOfType("MlCopilotClientMessage", &schema, &openapi3.T{Components: &openapi3.Components{Schemas: openapi3.Schemas{}}}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(data.Types["MlCopilotMessageUser"], "Content string") || !strings.Contains(data.Types["MlCopilotMessagePing"], "Type string") {
			t.Fatalf("payload growth changed variant names or fields: %v", data.Types)
		}
		if schema.OneOf[0].Value.Properties["type"] == nil {
			t.Fatal("generation removed the wire discriminator from the input schema")
		}
	}
}

func TestCopilotVariantNamesRejectAmbiguousTags(t *testing.T) {
	for _, input := range []string{
		`{"oneOf":[{"type":"object","properties":{}}]}`,
		`{"oneOf":[{"type":"object","properties":{"type":{"enum":["user"]}}},{"type":"object","properties":{"type":{"enum":["user"]}}}]}`,
	} {
		var schema openapi3.Schema
		if err := json.Unmarshal([]byte(input), &schema); err != nil {
			t.Fatal(err)
		}
		data := Data{Types: map[string]string{}}
		if err := data.generateOneOfType("MlCopilotClientMessage", &schema, &openapi3.T{Components: &openapi3.Components{Schemas: openapi3.Schemas{}}}); err == nil {
			t.Fatal("expected an error for a missing or duplicate tag")
		}
	}
}
