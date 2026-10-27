package cloudflare

import (
	"encoding/pem"
	"testing"
)

func TestParseOriginCert(t *testing.T) {
	block := func(typ, body string) []byte {
		return pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: []byte(body)})
	}
	token := block("ARGO TUNNEL TOKEN", `{"zoneID":"z","accountID":"a","apiToken":"t"}`)

	cert, err := parseOriginCert(append(block("PRIVATE KEY", "legacy"), token...))
	if err != nil {
		t.Fatal(err)
	}
	if cert != (OriginCert{ZoneID: "z", AccountID: "a", APIToken: "t"}) {
		t.Errorf("got %+v", cert)
	}

	for name, data := range map[string][]byte{
		"no token":      block("CERTIFICATE", ""),
		"two tokens":    append(token, token...),
		"unknown block": append(token, block("FOO", "")...),
		"fedramp":       block("ARGO TUNNEL TOKEN", `{"accountID":"a","apiToken":"t","endpoint":"FED"}`),
	} {
		if _, err := parseOriginCert(data); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}
