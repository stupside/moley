// Package framework provides resource lifecycle management.
package framework

import (
	"context"

	"github.com/stupside/moley/internal/core/domain"
)

type ResourceHandler interface {
	Name(ctx context.Context) string
	Up(ctx context.Context, payload any) error
	Down(ctx context.Context, payload any) error
	Status(ctx context.Context, payload any) (domain.State, error)
}
