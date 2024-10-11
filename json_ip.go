package kittycad

import (
	"bytes"
	"net/netip"
	"strings"
)

// IP is a wrapper around ip.IP which marshals to and from empty strings.
type IP struct {
	*netip.Addr
}

// MarshalJSON implements the json.Marshaler interface.
func (u IP) MarshalJSON() ([]byte, error) {
	return u.MarshalText()
}

func (u IP) String() string {
	return u.String()
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

	ip, err := netip.ParseAddr(strings.Trim(string(data), `"`))
	if err != nil {
		return err
	}
	*u = IP{&ip}
	return
}
