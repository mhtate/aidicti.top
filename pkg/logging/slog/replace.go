package slog

import "log/slog"

func AddSourceIndent(f func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		attr := f(groups, a)

		if (len(groups) != 0) && (groups[0] == "stack") && (attr.Key == "file") {
			attr.Value = slog.StringValue(" " + attr.Value.String() + " ")
		}

		return attr
	}
}

func ChangeTimeFormat(f func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		attr := f(groups, a)

		if len(groups) != 0 {
			return attr
		}

		if attr.Key == "time" {
		}

		return attr
	}
}
