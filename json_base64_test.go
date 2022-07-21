package kittycad

import (
	"testing"
)

func TestBase64(t *testing.T) {
	urlString := "aGVsbG8gd29ybGQK"
	var jsonBase64 Base64
	if err := jsonBase64.UnmarshalJSON([]byte(`"` + urlString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestBase64EmptyString(t *testing.T) {
	urlString := ""
	var jsonBase64 Base64
	if err := jsonBase64.UnmarshalJSON([]byte(`"` + urlString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestBase64Null(t *testing.T) {
	var jsonBase64 Base64
	if err := jsonBase64.UnmarshalJSON([]byte("null")); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if err := jsonBase64.UnmarshalJSON(nil); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}
