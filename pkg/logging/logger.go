package logging

import (
	"context"

	"github.com/go-kit/kit/log"

	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

// WithContext TODOC
func WithContext(ctx context.Context, logger log.Logger) log.Logger {
	rid, ok := ctx.Value(util.ContextKey("request_id")).(string)
	if !ok {
		return log.With(logger, "request_id", "unknown")
	}

	return log.With(logger, "request_id", rid)

}
