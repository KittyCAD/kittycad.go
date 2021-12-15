package kittycad

import (
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
	session, err := client.MetaDebugSession()
	if err != nil {
		t.Fatalf("getting the session failed: %v", err)
	}
	if session.ID == "" {
		t.Fatalf("the session ID is empty")
	}
}

func TestGetInstance(t *testing.T) {
	client := getClient(t)
	instance, err := client.MetaDebugInstance()
	if err != nil {
		t.Fatalf("getting the instance failed: %v", err)
	}
	if instance.ID == "" {
		t.Fatalf("the instance ID is empty")
	}
}

func TestPing(t *testing.T) {
	client := getClient(t)
	message, err := client.Ping()
	if err != nil {
		t.Fatalf("pinging the server failed: %v", err)
	}

	if message.Message != "pong" {
		t.Fatalf("the message is not pong: %v", message.Message)
	}
}
