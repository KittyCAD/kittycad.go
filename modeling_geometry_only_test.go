package kittycad

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestModelingGeometryIntent(t *testing.T) {
	for _, geometryOnly := range []bool{false, true} {
		t.Run(strconv.FormatBool(geometryOnly), func(t *testing.T) {
			queries := make(chan url.Values, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				queries <- r.URL.Query()
				upgrader := websocket.Upgrader{}
				connection, err := upgrader.Upgrade(w, r, nil)
				if err == nil {
					connection.Close()
				}
			}))
			defer server.Close()
			client, err := NewClient("test-token", "geometry-intent-test")
			if err != nil {
				t.Fatal(err)
			}
			if err := client.WithBaseURL(strings.Replace(server.URL, "http://", "ws://", 1)); err != nil {
				t.Fatal(err)
			}
			connection, err := client.Modeling.CommandsWs(0, 0, 0, false, "", false, "", false, "", "", false, geometryOnly, 0, nil)
			if err != nil {
				t.Fatal(err)
			}
			connection.Close()
			query := <-queries
			if query.Get("geometry_only") != strconv.FormatBool(geometryOnly) {
				t.Fatalf("geometry_only: %v", query)
			}
			if query.Get("webrtc") != "false" {
				t.Fatalf("webrtc: %v", query)
			}
			for _, unused := range []string{"post_effect", "video_res_width", "video_res_height"} {
				if query.Has(unused) {
					t.Fatalf("unexpected %s: %v", unused, query)
				}
			}
		})
	}
}
