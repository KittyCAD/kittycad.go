package kittycad

import (
	"bytes"
	"net"
)

// IP is a wrapper around ip.IP which marshals to and from empty strings.
type IP struct {
	*net.IP
}

// MarshalJSON implements the json.Marshaler interface.
func (u IP) MarshalJSON() ([]byte, error) {
	if u.IP == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + u.IP.String() + `"`), nil
}

func (u IP) String() string {
	if u.IP == nil {
		return ""
	}

	return u.IP.String()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (u *IP) UnmarshalJSON(data []byte) (err error) {
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

	var ip *net.IP = nil
	if err = ip.UnmarshalText(data); err != nil {
		return err
	}
	*u = IP{ip}
	return
}
