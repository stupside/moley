package tunnel

import "fmt"

// castPayload casts a payload to the expected pointer type T
func castPayload[T any](payload any) (*T, error) {
	if v, ok := payload.(*T); ok {
		return v, nil
	}
	var zero *T
	return nil, fmt.Errorf("invalid payload type: expected %T, got %T", zero, payload)
}
