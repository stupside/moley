package steps

import (
	"context"
)

// Step represents a deployable/cleanupable step
type Step interface {
	// Name returns the name of the step
	Name() string
	// Up deploys the step
	Up(ctx context.Context) error
	// Down cleans up the step
	Down(ctx context.Context) error
}
