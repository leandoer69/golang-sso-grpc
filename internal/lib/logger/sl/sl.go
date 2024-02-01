package sl

import "log/slog"

func Error(err error) slog.Attr {
	attr := slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}

	return attr
}
