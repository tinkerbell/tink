package clientctx

import (
	"context"

	"github.com/tinkerbell/tink/client"
)

type key struct{}

func Set(ctx context.Context, c *client.FullClient) context.Context {
	return context.WithValue(ctx, key{}, c)
}

func Get(ctx context.Context) *client.FullClient {
	c, ok := ctx.Value(key{}).(*client.FullClient)
	if !ok {
		return nil
	}
	return c
}
