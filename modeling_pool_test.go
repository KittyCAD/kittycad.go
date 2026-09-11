package kittycad

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestModelingPoolSelection(t *testing.T) {
	for _, pool := range []string{"cpu", "default", ""} {
		t.Run("pool="+pool, func(t *testing.T) {
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
			connection, err := client.Modeling.CommandsWs(0, 0, 0, false, "", false, pool, false, "", "", false, 0, nil)
			if err != nil {
				t.Fatal(err)
			}
			connection.Close()
			query := <-queries
			if query.Get("pool") != pool {
				t.Fatalf("pool: %v", query)
			}
			if query.Get("webrtc") != "false" {
				t.Fatalf("webrtc: %v", query)
			}
			for _, unused := range []string{"geometry_only", "post_effect", "video_res_width", "video_res_height"} {
				if query.Has(unused) {
					t.Fatalf("unexpected %s: %v", unused, query)
				}
			}
		})
	}
}
