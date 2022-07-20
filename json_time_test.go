package kittycad

import (
	"testing"
)

func TestTime(t *testing.T) {
	timeString := "2018-01-01T00:00:00Z"
	var jsonTime Time
	if err := jsonTime.UnmarshalJSON([]byte(`"` + timeString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestTimeEmptyString(t *testing.T) {
	timeString := ""
	var jsonTime Time
	if err := jsonTime.UnmarshalJSON([]byte(`"` + timeString + `"`)); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}

func TestTimeNull(t *testing.T) {
	var jsonTime Time
	if err := jsonTime.UnmarshalJSON([]byte("null")); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if err := jsonTime.UnmarshalJSON(nil); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
}
