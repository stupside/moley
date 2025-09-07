package framework

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Resource struct {
	Handler string `json:"handler"`
	Payload any    `json:"payload"`
}

func NewResource(ctx context.Context, handler ResourceHandler, payload any) (*Resource, error) {
	return &Resource{
		Payload: payload,
		Handler: handler.Name(ctx),
	}, nil
}

// Hash returns a deterministic hash of the payload using JSON serialization.
func (r *Resource) Hash() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to hash payload: %w", err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}
