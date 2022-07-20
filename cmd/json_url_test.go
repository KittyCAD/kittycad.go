package main

import (
	"testing"
)

func TestURL(t *testing.T) {
	urlString := "https://www.google.com"
	var jsonURL URL
	if err := jsonURL.UnmarshalJSON([]byte(`"` + urlString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestURLEmptyString(t *testing.T) {
	urlString := ""
	var jsonURL URL
	if err := jsonURL.UnmarshalJSON([]byte(`"` + urlString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestURLNull(t *testing.T) {
	var jsonURL URL
	if err := jsonURL.UnmarshalJSON([]byte("null")); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if err := jsonURL.UnmarshalJSON(nil); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}
