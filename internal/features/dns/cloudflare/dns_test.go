package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cfgo "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

// TestRouteRecord checks the v6 request shapes against a fake API: exact-match
// record lookup and a typed CNAME body.
func TestRouteRecord(t *testing.T) {
	var created map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		result := "[]"
		switch {
		case r.URL.Path == "/zones" && q.Get("page") == "":
			result = `[{"id":"zone1"}]`
		case r.URL.Path == "/zones/zone1/dns_records" && r.Method == http.MethodGet:
			if q.Get("name.exact") != "app.example.com" || q.Get("content.exact") != "uuid.cfargotunnel.com" {
				t.Errorf("unexpected record query: %s", r.URL.RawQuery)
			}
		case r.URL.Path == "/zones/zone1/dns_records" && r.Method == http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
				t.Error(err)
			}
			result = `{"id":"rec1"}`
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":true,"errors":[],"messages":[],"result":%s}`, result)
	}))
	defer srv.Close()

	client := cfgo.NewClient(option.WithAPIToken("test"), option.WithBaseURL(srv.URL))
	if err := NewDNSService(client, false).RouteRecord(t.Context(), "uuid", "example.com", "app"); err != nil {
		t.Fatal(err)
	}

	want := map[string]any{"name": "app.example.com", "type": "CNAME", "content": "uuid.cfargotunnel.com", "proxied": true, "ttl": float64(1)}
	for k, v := range want {
		if created[k] != v {
			t.Errorf("created[%q] = %v, want %v (body: %v)", k, created[k], v, created)
		}
	}
}
