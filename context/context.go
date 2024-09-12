package context

import "context"

var Ctx context.Context

func init() {
	Ctx = context.Background()
}
