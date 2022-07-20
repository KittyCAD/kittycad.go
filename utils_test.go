package kittycad

import (
	"net/url"
	"testing"
)

type expandTest struct {
	in         string
	expansions map[string]string
	want       string
}

func TestExpandURL(t *testing.T) {
	var testServerURL = "http://example.com"

	var expandTests = []expandTest{
		// no expansions
		{
			"",
			map[string]string{},
			testServerURL,
		},
		// multiple expansions, no escaping
		{
			"file/convert/{{.srcFormat}}/{{.outputFormat}}",
			map[string]string{
				"srcFormat":    "step",
				"outputFormat": "obj",
			},
			testServerURL + "/file/convert/step/obj",
		},
		// Path params added as extras.
		{
			"file/convert",
			map[string]string{
				"srcFormat":    "step",
				"outputFormat": "obj",
			},
			testServerURL + "/file/convert?outputFormat=obj&srcFormat=step",
		},
	}

	for i, test := range expandTests {
		uri := resolveRelative(testServerURL, test.in)
		u, err := url.Parse(uri)
		if err != nil {
			t.Fatalf("parsing url %q failed: %v", test.in, err)
		}
		expandURL(u, test.expansions)
		got := u.String()
		if got != test.want {
			t.Errorf("got %q expected %q in test %d", got, test.want, i+1)
		}
	}
}
