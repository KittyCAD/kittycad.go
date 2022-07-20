package main

import (
	"testing"
)

func TestUUID(t *testing.T) {
	uuidString := "f8e50cf2-a25a-4e88-baaa-96cd5e551242"
	var jsonUUID UUID
	if err := jsonUUID.UnmarshalJSON([]byte(`"` + uuidString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestUUIDEmptyString(t *testing.T) {
	uuidString := ""
	var jsonUUID UUID
	if err := jsonUUID.UnmarshalJSON([]byte(`"` + uuidString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestUUIDNull(t *testing.T) {
	var jsonUUID UUID
	if err := jsonUUID.UnmarshalJSON([]byte("null")); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if err := jsonUUID.UnmarshalJSON(nil); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}
