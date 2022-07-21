package main

import (
	"bytes"
	"encoding/base64"
	"strings"
)

// Base64 is a wrapper around url.Base64 which marshals to and from empty strings.
type Base64 struct {
	Inner []byte
}

// MarshalJSON implements the json.Marshaler interface.
func (u Base64) MarshalJSON() ([]byte, error) {
	if u.Inner == nil || len(u.Inner) <= 0 {
		return []byte("null"), nil
	}

	return []byte(`"` + base64.StdEncoding.EncodeToString(u.Inner) + `"`), nil
}

func (u Base64) String() string {
	if u.Inner == nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(u.Inner)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (u *Base64) UnmarshalJSON(data []byte) (err error) {
	// By convention, unmarshalers implement UnmarshalJSON([]byte("null")) as a no-op.
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	if bytes.Equal(data, []byte("")) {
		return nil
	}

	if bytes.Equal(data, []byte(`""`)) {
		return nil
	}

	// Fractional seconds are handled implicitly by Parse.
	uu, err := base64.StdEncoding.DecodeString(strings.Trim(string(data), `"`))
	if err != nil {
		return err
	}
	*u = Base64{Inner: uu}
	return
}
