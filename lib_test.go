package kittycad

import (
	"os"
	"path/filepath"
	"testing"
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
