package types

import "context"

type Quota interface {
	Prepare(ctx context.Context, target string, opts map[string]string) error
	Setup(ctx context.Context, target string, size int, opts map[string]string) error
	Remove(ctx context.Context, target string) error
	Get(ctx context.Context, target string) (int error)
}
