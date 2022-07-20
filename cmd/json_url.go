package main

import (
	"bytes"
	"net/url"
)

// URL is a wrapper around url.URL which marshals to and from empty strings.
type URL struct {
	*url.URL
}

// MarshalJSON implements the json.Marshaler interface.
func (u URL) MarshalJSON() ([]byte, error) {
	if u.URL == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + u.URL.String() + `"`), nil
}

func (u URL) String() string {
	if u.URL == nil {
		return ""
	}

	return u.URL.String()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (u *URL) UnmarshalJSON(data []byte) (err error) {
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
	uu, err := url.Parse(string(data))
	if err != nil {
		return err
	}
	*u = URL{uu}
	return
}
