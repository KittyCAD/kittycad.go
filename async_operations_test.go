package kittycad

import (
	"fmt"
	"testing"
	"time"
)

func createTestAsyncFileConversion(t *testing.T, client *Client) *FileConversion {
	t.Helper()

	coords := System{
		Forward: AxisDirectionPair{Axis: AxiY, Direction: DirectionNegative},
		Up:      AxisDirectionPair{Axis: AxiZ, Direction: DirectionPositive},
	}
	form := NewMultipartForm()
	if err := form.WriteJSONField("body", ConversionParams{
		SrcFormat:    map[string]any{"type": FileImportFormatStl, "coords": coords, "units": UnitLengthMm},
		OutputFormat: map[string]any{"type": FileExportFormatObj, "coords": coords, "units": UnitLengthMm},
	}); err != nil {
		t.Fatalf("writing the conversion parameters failed: %v", err)
	}

	body := []byte(fmt.Sprintf("solid kittycad-go-%d\nfacet normal 0 0 1\nouter loop\nvertex 0 0 0\nvertex 1 0 0\nvertex 0 1 0\nendloop\nendfacet\nendsolid\n", time.Now().UnixNano()))
	if err := form.WriteFilePart("source", "source.stl", "text/plain", body); err != nil {
		t.Fatalf("writing the STL attachment failed: %v", err)
	}

	created, err := client.File.CreateConversionOptions(form)
	if err != nil {
		t.Fatalf("creating the async file conversion failed: %v", err)
	}
	if created.ID.String() == "" {
		t.Fatalf("the async file conversion ID is empty")
	}

	return created
}

func getAsyncOperationResultMap(t *testing.T, result *any) map[string]any {
	t.Helper()

	if result == nil {
		t.Fatalf("the async operation result is nil")
	}

	resultMap, ok := (*result).(map[string]any)
	if !ok {
		t.Fatalf("the async operation result has unexpected type: %T", *result)
	}

	return resultMap
}

func getAsyncOperationStringField(t *testing.T, resultMap map[string]any, field string) string {
	t.Helper()

	value, ok := resultMap[field].(string)
	if !ok || value == "" {
		t.Fatalf("the async operation result is missing %q: %#v", field, resultMap)
	}

	return value
}

func TestAsyncOperationStatus(t *testing.T) {
	client := getClient(t)
	created := createTestAsyncFileConversion(t, client)

	deadline := time.Now().Add(3 * time.Minute)
	for {
		result, err := client.APICall.GetAsyncOperation(created.ID)
		if err != nil {
			t.Fatalf("getting the async operation failed: %v", err)
		}

		resultMap := getAsyncOperationResultMap(t, result)

		gotID := getAsyncOperationStringField(t, resultMap, "id")
		if gotID != created.ID.String() {
			t.Fatalf("the async operation ID mismatch, got %q want %q", gotID, created.ID.String())
		}

		status := APICallStatus(getAsyncOperationStringField(t, resultMap, "status"))
		switch status {
		case APICallStatusCompleted:
			outputs, ok := resultMap["outputs"].(map[string]any)
			if !ok {
				t.Fatalf("the completed async operation result is missing outputs: %#v", resultMap)
			}
			if len(outputs) == 0 {
				t.Fatalf("the completed async operation result has no outputs: %#v", resultMap)
			}
			return
		case APICallStatusFailed:
			t.Fatalf("the async operation failed: %s", getAsyncOperationStringField(t, resultMap, "error"))
		case APICallStatusQueued, APICallStatusUploaded, APICallStatusInProgress:
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for the async operation to complete, last status: %s", status)
			}
			time.Sleep(2 * time.Second)
		default:
			t.Fatalf("unexpected async operation status: %s", status)
		}
	}
}
