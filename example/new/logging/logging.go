package logging

import (
	"errors"
	"fmt"
	"log/slog"
)

type ErrorWithAttrs struct {
	attrs []slog.Attr
	err   error
}

func (e *ErrorWithAttrs) Error() string {
	return e.err.Error()
}

func (e *ErrorWithAttrs) Unwrap() error {
	return e.err
}

func (e *ErrorWithAttrs) Attrs() []slog.Attr {
	return e.attrs
}

func Errorf(format string, args ...any) error {
	var attrs []slog.Attr
	var filteredArgs []any

	for _, arg := range args {
		if attr, ok := arg.(slog.Attr); ok {
			attrs = append(attrs, attr)
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	return &ErrorWithAttrs{
		attrs: attrs,
		err:   fmt.Errorf(format, filteredArgs...),
	}
}

func AttrsFromError(err error) []slog.Attr {
	var ewa *ErrorWithAttrs
	if errors.As(err, &ewa) {
		return ewa.Attrs()
	}
	return nil
}

func Error(err error) slog.Attr {
	attrs := AttrsFromError(err)
	if len(attrs) == 0 {
		return slog.String("error", err.Error())
	}
	attrs = append(attrs, slog.String("error", err.Error()))
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}

	return slog.Group("", args...)
}
