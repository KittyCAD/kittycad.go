package kittycad

import (
	"encoding/json"
	"testing"
)

func TestFileConversionWithEmptyCompletedAt(t *testing.T) {
	j := `{"completed_at":"","created_at":"2022-02-03T00:04:01Z","id":"845eebb5-f45d-4273-97bd-d5c7e398f7e0","output":"","output_format":"obj","src_format":"stl","started_at":"","status":"Uploaded"}`
	var f FileConversion
	err := json.Unmarshal([]byte(j), &f)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}
}
