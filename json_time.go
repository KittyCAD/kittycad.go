package kittycad

import (
	"bytes"
	"time"
)

// JSONTime is a wrapper around time.Time which marshals to and from empty strings.
type JSONTime struct {
	*time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(t.Format("\"" + time.RFC3339 + "\"")), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {

	// by convention, unmarshalers implement UnmarshalJSON([]byte("null")) as a no-op.
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	if bytes.Equal(data, []byte("")) {
		return nil
	}

	if bytes.Equal(data, []byte("\"\"")) {
		return nil
	}

	// Fractional seconds are handled implicitly by Parse.
	tt, err := time.Parse("\""+time.RFC3339+"\"", string(data))
	*t = JSONTime{&tt}
	return
}
