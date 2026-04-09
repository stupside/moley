package domain

import (
	"encoding/json"
	"maps"
)

// Policy is a named, reusable Cloudflare Access policy.
// Name is the reference key; all other fields are CF API fields passed through as-is.
type Policy struct {
	Name  string         `yaml:"name"    json:"name"  validate:"required"`
	Extra map[string]any `yaml:",remain" json:"-"`
}

// MarshalJSON inlines Extra so the full CF policy body is produced on serialization.
func (p Policy) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(p.Extra)+1)
	maps.Copy(m, p.Extra)
	m["name"] = p.Name
	return json.Marshal(m)
}

type Access struct {
	Policies []Policy `yaml:"policies,omitempty"`
}

func (a *Access) HasPolicies() bool {
	return a != nil && len(a.Policies) > 0
}

func (a *Access) PolicyByName(name string) (Policy, bool) {
	if a == nil {
		return Policy{}, false
	}
	for _, p := range a.Policies {
		if p.Name == name {
			return p, true
		}
	}
	return Policy{}, false
}
