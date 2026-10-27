package cloudflare

import (
	"reflect"
	"testing"

	"github.com/stupside/moley/v2/internal/domain"
	tunnelusecase "github.com/stupside/moley/v2/internal/features/tunnel/usecase"
)

func TestBuildIngress(t *testing.T) {
	web := domain.TargetConfig{Hostname: "localhost", Port: 3000, Protocol: domain.ProtocolHTTP,
		OriginRequest: &domain.OriginRequest{HTTPHostHeader: "localhost"}}
	api := domain.TargetConfig{Hostname: "localhost", Port: 8080, Protocol: domain.ProtocolHTTP}

	rules := buildIngress(&domain.Ingress{Zone: "example.com", Apps: []domain.AppConfig{
		{Target: web, Expose: domain.ExposeConfig{Subdomain: "app"}},
		{Target: api, Expose: domain.ExposeConfig{Subdomain: "app", Path: "/api"}},
	}}, map[string]tunnelusecase.AccessGuard{
		"app.example.com/api": {TeamName: "team", AUD: "aud1"},
	})

	want := []ingressRule{
		{Hostname: "app.example.com", Path: `^/api(/|$)`, Service: "http://localhost:8080",
			OriginRequest: &originRequest{Access: &originAccess{Required: true, TeamName: "team", AudTag: []string{"aud1"}}}},
		{Hostname: "app.example.com", Service: "http://localhost:3000",
			OriginRequest: &originRequest{OriginRequest: *web.OriginRequest}},
		{Service: "http_status:404"},
	}
	if !reflect.DeepEqual(rules, want) {
		t.Errorf("buildIngress() =\n%#v\nwant\n%#v", rules, want)
	}
}
