package config

import (
	"reflect"
	"testing"
)

func TestNumericKeysToSliceHook_SequentialKeys(t *testing.T) {
	hook := numericKeysToSliceHookFunc()

	input := map[string]any{
		"0": map[string]any{"name": "first"},
		"1": map[string]any{"name": "second"},
		"2": map[string]any{"name": "third"},
	}

	result, err := hook.(func(reflect.Type, reflect.Type, any) (any, error))(
		reflect.TypeOf(input),
		reflect.TypeFor[[]any](),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slice, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(slice) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(slice))
	}

	// Verify order
	for i, expected := range []string{"first", "second", "third"} {
		m, ok := slice[i].(map[string]any)
		if !ok {
			t.Fatalf("element %d: expected map, got %T", i, slice[i])
		}
		if m["name"] != expected {
			t.Fatalf("element %d: expected name=%s, got %v", i, expected, m["name"])
		}
	}
}

func TestNumericKeysToSliceHook_NonNumericPassthrough(t *testing.T) {
	hook := numericKeysToSliceHookFunc()

	input := map[string]any{
		"host": "localhost",
		"port": 3000,
	}

	result, err := hook.(func(reflect.Type, reflect.Type, any) (any, error))(
		reflect.TypeOf(input),
		reflect.TypeFor[[]any](),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return input unchanged (non-numeric keys)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map passthrough, got %T", result)
	}
	if m["host"] != "localhost" {
		t.Fatalf("expected host=localhost, got %v", m["host"])
	}
}

func TestNumericKeysToSliceHook_NonSequentialPassthrough(t *testing.T) {
	hook := numericKeysToSliceHookFunc()

	input := map[string]any{
		"0": "first",
		"2": "third", // gap: no "1"
	}

	result, err := hook.(func(reflect.Type, reflect.Type, any) (any, error))(
		reflect.TypeOf(input),
		reflect.TypeFor[[]any](),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return input unchanged (non-sequential keys)
	_, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map passthrough for non-sequential keys, got %T", result)
	}
}

func TestNumericKeysToSliceHook_NonSliceTarget(t *testing.T) {
	hook := numericKeysToSliceHookFunc()

	input := map[string]any{
		"0": "first",
		"1": "second",
	}

	// Target is a struct, not a slice — should pass through
	type Dummy struct{}
	result, err := hook.(func(reflect.Type, reflect.Type, any) (any, error))(
		reflect.TypeOf(input),
		reflect.TypeFor[Dummy](),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map passthrough for non-slice target, got %T", result)
	}
}
