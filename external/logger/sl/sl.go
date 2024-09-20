package sl

import (
	"log/slog"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func String(key string, value string) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(value),
	}
}

func Any(key string, value interface{}) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.AnyValue(value),
	}
}

func Bool(key string, value bool) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.BoolValue(value),
	}
}

func Float64(key string, value float64) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.Float64Value(value),
	}
}

func Int(key string, value int) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.IntValue(value),
	}
}

func Int64(key string, value int64) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.Int64Value(value),
	}
}
