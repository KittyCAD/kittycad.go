package main

import (
	"bytes"
	"time"
)

// Time is a wrapper around time.Time which marshals to and from empty strings.
type Time struct {
	*time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time == nil {
		return []byte("null"), nil
	}

	return []byte(t.Format(`"` + time.RFC3339 + `"`)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Time) UnmarshalJSON(data []byte) (err error) {

	// by convention, unmarshalers implement UnmarshalJSON([]byte("null")) as a no-op.
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
	tt, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
	if err != nil {
		return err
	}

	*t = Time{&tt}
	return
}
