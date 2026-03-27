package config

import (
	"reflect"
	"testing"
)

func TestNumericKeysToSliceHook_SequentialKeys(t *testing.T) {
	hook := numericKeysToSliceHookFunc()

	input := map[string]interface{}{
		"0": map[string]interface{}{"name": "first"},
		"1": map[string]interface{}{"name": "second"},
		"2": map[string]interface{}{"name": "third"},
	}

	result, err := hook.(func(reflect.Type, reflect.Type, interface{}) (interface{}, error))(
		reflect.TypeOf(input),
		reflect.TypeOf([]interface{}{}),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slice, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(slice) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(slice))
	}

	// Verify order
	for i, expected := range []string{"first", "second", "third"} {
		m, ok := slice[i].(map[string]interface{})
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

	input := map[string]interface{}{
		"host": "localhost",
		"port": 3000,
	}

	result, err := hook.(func(reflect.Type, reflect.Type, interface{}) (interface{}, error))(
		reflect.TypeOf(input),
		reflect.TypeOf([]interface{}{}),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return input unchanged (non-numeric keys)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map passthrough, got %T", result)
	}
	if m["host"] != "localhost" {
		t.Fatalf("expected host=localhost, got %v", m["host"])
	}
}

func TestNumericKeysToSliceHook_NonSequentialPassthrough(t *testing.T) {
	hook := numericKeysToSliceHookFunc()

	input := map[string]interface{}{
		"0": "first",
		"2": "third", // gap: no "1"
	}

	result, err := hook.(func(reflect.Type, reflect.Type, interface{}) (interface{}, error))(
		reflect.TypeOf(input),
		reflect.TypeOf([]interface{}{}),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return input unchanged (non-sequential keys)
	_, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map passthrough for non-sequential keys, got %T", result)
	}
}

func TestNumericKeysToSliceHook_NonSliceTarget(t *testing.T) {
	hook := numericKeysToSliceHookFunc()

	input := map[string]interface{}{
		"0": "first",
		"1": "second",
	}

	// Target is a struct, not a slice — should pass through
	type Dummy struct{}
	result, err := hook.(func(reflect.Type, reflect.Type, interface{}) (interface{}, error))(
		reflect.TypeOf(input),
		reflect.TypeOf(Dummy{}),
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map passthrough for non-slice target, got %T", result)
	}
}
