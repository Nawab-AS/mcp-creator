package backend

import (
	"context"
)

type Prebuilt struct{ ctx context.Context }

func (a *Prebuilt) Startup(ctx context.Context) { a.ctx = ctx }
