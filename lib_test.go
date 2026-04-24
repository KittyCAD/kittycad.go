package kittycad

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func getClient(t *testing.T) *Client {
	t.Helper()

	if os.Getenv(TokenEnvVar) == "" && os.Getenv("KITTYCAD_API_TOKEN") == "" {
		t.Skipf("skipping integration test: set %s or KITTYCAD_API_TOKEN", TokenEnvVar)
	}

	client, err := NewClientFromEnv("kittycad.go/tests")
	if err != nil {
		t.Fatalf("creating the client failed: %v", err)
	}
	return client
}

func TestGetSession(t *testing.T) {
	client := getClient(t)
	_, err := client.User.GetSelf()
	if err != nil {
		t.Fatalf("getting the session failed: %v", err)
	}
}

func TestPing(t *testing.T) {
	client := getClient(t)
	message, err := client.Meta.Ping()
	if err != nil {
		t.Fatalf("pinging the server failed: %v", err)
	}

	if message.Message != "pong" {
		t.Fatalf("the message is not pong: %v", message.Message)
	}
}

func getTestFileConversionBody(t *testing.T) []byte {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting the current working directory failed: %v", err)
	}

	file := filepath.Join(cwd, "assets", "testing.stl")
	body, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("reading the test file %q failed: %v", file, err)
	}

	return body
}

func createTestFileConversion(t *testing.T, client *Client) *FileConversion {
	t.Helper()

	fc, err := client.File.CreateConversion(FileImportFormatStl, FileExportFormatObj, getTestFileConversionBody(t))
	if err != nil {
		t.Fatalf("getting the file conversion failed: %v", err)
	}

	if fc.ID.String() == "" {
		t.Fatalf("the file conversion ID is empty")
	}

	return fc
}

func createTestTextToCadMultiFileIteration(t *testing.T, client *Client) *TextToCadMultiFileIteration {
	t.Helper()

	form := NewMultipartForm()

	if err := form.WriteJSONField("body", TextToCadMultiFileIterationBody{
		KclVersion:   "1.0",
		ProjectName:  "kittycad.go async operation test",
		Prompt:       "Add a simple cube to main.kcl and a cylinder to subdir/main.kcl",
		SourceRanges: []SourceRangePrompt{},
	}); err != nil {
		t.Fatalf("writing the multipart JSON body failed: %v", err)
	}

	if err := form.WriteFilePart("main.kcl", "main.kcl", "text/plain", []byte("// Glorious cube\n\nsideLength = 10\n")); err != nil {
		t.Fatalf("writing the main.kcl attachment failed: %v", err)
	}

	if err := form.WriteFilePart("subdir/main.kcl", "subdir/main.kcl", "text/plain", []byte("// Glorious cylinder\n\nheight = 20\n")); err != nil {
		t.Fatalf("writing the subdir/main.kcl attachment failed: %v", err)
	}

	created, err := client.Ml.CreateTextToCadMultiFileIteration(form)
	if err != nil {
		t.Fatalf("creating the async text-to-cad multi-file iteration failed: %v", err)
	}

	if created.ID.String() == "" {
		t.Fatalf("the async text-to-cad multi-file iteration ID is empty")
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

func TestFileConversion(t *testing.T) {
	client := getClient(t)

	fc := createTestFileConversion(t, client)

	if fc.Status != "completed" {
		t.Fatalf("the file conversion status is not `completed`: %v", fc.Status)
	}

	// Make sure we have a started at time.
	if fc.StartedAt.IsZero() {
		t.Fatalf("the file conversion started at time is zero")
	}

	if fc.CompletedAt.IsZero() {
		t.Fatalf("the file conversion completed at time is zero")
	}

	if len(fc.Outputs) == 0 {
		t.Fatalf("the file conversion output is empty")
	}

	for _, output := range fc.Outputs {
		if len(output.Inner) == 0 {
			t.Fatalf("the file conversion output body is empty")
		}
	}
}

func TestAsyncOperationStatus(t *testing.T) {
	client := getClient(t)
	created := createTestTextToCadMultiFileIteration(t, client)

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
			if _, ok := outputs["main.kcl"]; !ok {
				t.Fatalf("the completed async operation result is missing main.kcl: %#v", outputs)
			}
			if _, ok := outputs["subdir/main.kcl"]; !ok {
				t.Fatalf("the completed async operation result is missing subdir/main.kcl: %#v", outputs)
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
