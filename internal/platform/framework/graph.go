package framework

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// node is the type-erased interface used by the reconciler's DAG.
type node interface {
	name() string
	dependencies() []string
	resolve(reg *OutputRegistry) error
	reconcile(ctx context.Context, lf *LockFile) error
	stop(ctx context.Context, lf *LockFile) error
}

// OutputRegistry holds outputs keyed by handler name + resource key.
type OutputRegistry struct {
	data map[string]any // key: "handlerName/resourceKey" → output value
}

func newOutputRegistry() *OutputRegistry {
	return &OutputRegistry{data: make(map[string]any)}
}

func (r *OutputRegistry) set(handlerName, key string, output any) {
	r.data[handlerName+"/"+key] = output
}

func (r *OutputRegistry) get(handlerName, key string) (any, bool) {
	v, ok := r.data[handlerName+"/"+key]
	return v, ok
}

// GetOutput retrieves a typed output from the registry.
// Returns zero value if not found or type mismatch.
func GetOutput[T any](reg *OutputRegistry, handlerName string, key string) (T, bool) {
	v, ok := reg.get(handlerName, key)
	if !ok {
		var zero T
		return zero, false
	}

	// Try direct type assertion first
	if typed, ok := v.(T); ok {
		return typed, true
	}

	// Fall back to JSON round-trip for interface{} → concrete type
	var result T
	if err := unmarshalData(v, &result); err != nil {
		return result, false
	}
	return result, true
}

// computeHash returns a sha256 hex digest of the JSON-serialized input.
func computeHash(input any) (string, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to hash input: %w", err)
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

// Snapshot stores both input and output for a lock entry.
type Snapshot[TInput any, TOutput any] struct {
	Input  TInput  `json:"input"`
	Output TOutput `json:"output"`
}

// InputResolver resolves inputs for a handler using outputs from upstream handlers.
type InputResolver[TInput any] func(reg *OutputRegistry) ([]TInput, error)
