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

func TestBase64String(t *testing.T) {
	var jsonBase64 Base64
	if err := jsonBase64.UnmarshalJSON([]byte("diAtMC4wMDEgLTAuMDAxIDAuMDAxCnYgMC4wMDEgLTAuMDAxIDAuMDAxCnYgMC4wMDEgMC4wMDEgMC4wMDEKdiAtMC4wMDEgMC4wMDEgMC4wMDEKdiAtMC4wMDEgMC4wMDEgLTAuMDAxCnYgMC4wMDEgMC4wMDEgLTAuMDAxCnYgMC4wMDEgLTAuMDAxIC0wLjAwMQp2IC0wLjAwMSAtMC4wMDEgLTAuMDAxCnZuIDAgMCAxCnZuIDAgMSAwCnZuIDEgMCAwCnZuIDAgMCAtMQp2biAtMSAwIDAKdm4gMCAtMSAwCm8gVW5uYW1lZC0wCmYgMS8vMSAyLy8xIDMvLzEKZiAxLy8xIDMvLzEgNC8vMQpmIDUvLzIgNC8vMiAzLy8yCmYgNS8vMiAzLy8yIDYvLzIKZiA2Ly8zIDMvLzMgMi8vMwpmIDYvLzMgMi8vMyA3Ly8zCmYgNy8vNCA4Ly80IDUvLzQKZiA3Ly80IDUvLzQgNi8vNApmIDgvLzUgMS8vNSA0Ly81CmYgOC8vNSA0Ly81IDUvLzUKZiA3Ly82IDIvLzYgMS8vNgpmIDcvLzYgMS8vNiA4Ly82Cg")); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if err := jsonBase64.UnmarshalJSON(nil); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}
