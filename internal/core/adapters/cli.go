package adapters

import (
	"context"
)

type CliServer interface {
	Handler(ctx context.Context, url string) (*[]string, error)
}
