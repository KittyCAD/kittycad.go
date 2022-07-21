package kittycad

import (
	"bytes"

	"github.com/google/uuid"
)

// UUID is a wrapper around uuid.UUID which marshals to and from empty strings.
type UUID struct {
	*uuid.UUID
}

// MarshalJSON implements the json.Marshaler interface.
func (u UUID) MarshalJSON() ([]byte, error) {
	if u.UUID == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + u.UUID.String() + `"`), nil
}

func (u UUID) String() string {
	if u.UUID == nil {
		return ""
	}

	return u.UUID.String()
}

// ParseUUID parses a UUID from a string.
func ParseUUID(s string) UUID {
	u, _ := uuid.Parse(s)
	return UUID{&u}
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (u *UUID) UnmarshalJSON(data []byte) (err error) {
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
	uu, err := uuid.Parse(string(data))
	if err != nil {
		return err
	}
	*u = UUID{&uu}
	return
}
