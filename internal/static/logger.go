// Package zap provides a logger that writes to a go.uber.org/zap.Logger.
package static

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jackc/pgx/v5/tracelog"
)

type Logger struct {
	name string
	ctx  context.Context
}

func NewLogger(ctx context.Context, name string) *Logger {
	return &Logger{name: name, ctx: tflog.NewSubsystem(ctx, name)}
}

func (pl *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	switch level {
	case tracelog.LogLevelTrace:
		tflog.SubsystemTrace(pl.ctx, pl.name, msg, data)
	case tracelog.LogLevelDebug:
		tflog.SubsystemDebug(pl.ctx, pl.name, msg, data)
	case tracelog.LogLevelInfo:
		tflog.SubsystemInfo(pl.ctx, pl.name, msg, data)
	case tracelog.LogLevelWarn:
		tflog.SubsystemWarn(pl.ctx, pl.name, msg, data)
	case tracelog.LogLevelError:
		tflog.SubsystemError(pl.ctx, pl.name, msg, data)
	default:
		tflog.SubsystemError(pl.ctx, pl.name, msg, data)
	}
}
