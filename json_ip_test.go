package kittycad

import (
	"testing"
)

func TestIP(t *testing.T) {
	ipString := "192.158.1.38"
	var jsonIP IP
	if err := jsonIP.UnmarshalJSON([]byte(`"` + ipString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestIPEmptyString(t *testing.T) {
	ipString := ""
	var jsonIP IP
	if err := jsonIP.UnmarshalJSON([]byte(`"` + ipString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestIPNull(t *testing.T) {
	var jsonIP IP
	if err := jsonIP.UnmarshalJSON([]byte("null")); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if err := jsonIP.UnmarshalJSON(nil); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}
