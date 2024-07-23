package cache

import (
	"context"
)

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) bool
	Set(ctx context.Context, key string, value interface{}) error
}
