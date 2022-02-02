package kittycad

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func getClient(t *testing.T) *Client {
	client, err := NewClientFromEnv("kittycad.go/tests")
	if err != nil {
		t.Fatalf("creating the client failed: %v", err)
	}
	return client
}

func TestGetSession(t *testing.T) {
	client := getClient(t)
	session, err := client.Meta.DebugSession()
	if err != nil {
		t.Fatalf("getting the session failed: %v", err)
	}
	if session.ID == "" {
		t.Fatalf("the session ID is empty")
	}
}

func TestGetInstance(t *testing.T) {
	client := getClient(t)
	instance, err := client.Meta.DebugInstance()
	if err != nil {
		t.Fatalf("getting the instance failed: %v", err)
	}
	if instance.ID == "" {
		t.Fatalf("the instance ID is empty")
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

func TestFileConversion(t *testing.T) {
	client := getClient(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting the current working directory failed: %v", err)
	}

	file := filepath.Join(cwd, "assets", "testing.stl")
	body, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("reading the test file %q failed: %v", file, err)
	}

	fc, output, err := client.File.ConvertWithBase64Helper(ValidSourceFileTypeStl, ValidOutputFileTypeObj, body)
	if err != nil {
		t.Fatalf("getting the file conversion failed: %v", err)
	}

	if fc.ID == "" {
		t.Fatalf("the file conversion ID is empty")
	}

	if fc.Status != "Completed" {
		t.Fatalf("the file conversion status is not `Completed`: %v", fc.Status)
	}

	if len(output) == 0 {
		t.Fatalf("the file conversion output is empty")
	}
}
