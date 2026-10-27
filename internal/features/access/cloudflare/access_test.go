package cloudflare

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cfgo "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

func TestTeamName(t *testing.T) {
	for _, tc := range []struct {
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{http.StatusOK, `{"success":true,"errors":[],"messages":[],"result":{"auth_domain":"acme.cloudflareaccess.com"}}`, "acme", false},
		{http.StatusForbidden, `{"success":false,"errors":[{"code":10000,"message":"Authentication error"}],"messages":[],"result":null}`, "", true},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(tc.status)
			fmt.Fprint(w, tc.body)
		}))
		client := cfgo.NewClient(option.WithAPIToken("test"), option.WithBaseURL(srv.URL), option.WithMaxRetries(0))
		got, err := NewAccessService(client, "acc", false).TeamName(t.Context())
		srv.Close()
		if got != tc.want || (err != nil) != tc.wantErr {
			t.Errorf("status %d: got (%q, %v), want (%q, err=%v)", tc.status, got, err, tc.want, tc.wantErr)
		}
	}
}
