package loadgen

import (
	"context"
	"time"
)

// Interface defines a set of methods for database load tests.
type Interface interface {
	Name() string
	Run(ctx context.Context) error
	RunWithInterval(ctx context.Context, interval time.Duration) error
}
