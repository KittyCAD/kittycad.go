package main

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func schemaTypeIncludes(schema *openapi3.Schema, typ string) bool {
	return schema != nil && schema.Type != nil && schema.Type.Includes(typ)
}

func schemaRefTypeIncludes(ref *openapi3.SchemaRef, typ string) bool {
	return ref != nil && ref.Value != nil && schemaTypeIncludes(ref.Value, typ)
}

func schemaTypes(typs ...string) *openapi3.Types {
	if len(typs) == 0 {
		return nil
	}

	types := openapi3.Types(typs)
	return &types
}

func schemaTypeString(schema *openapi3.Schema) string {
	if schema == nil || schema.Type == nil {
		return ""
	}

	return strings.Join(schema.Type.Slice(), "|")
}

func pathItems(paths *openapi3.Paths) map[string]*openapi3.PathItem {
	if paths == nil {
		return map[string]*openapi3.PathItem{}
	}

	return paths.Map()
}

func responseRefs(responses *openapi3.Responses) map[string]*openapi3.ResponseRef {
	if responses == nil {
		return map[string]*openapi3.ResponseRef{}
	}

	return responses.Map()
}
