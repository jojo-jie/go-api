package logging

import (
	"errors"
	"fmt"
	"log/slog"
)

// ErrorWithAttrs wraps an error with structured attributes.
type ErrorWithAttrs struct {
	Err   error
	attrs []slog.Attr // 字段改名为 attrs（小写开头）
}

func (e *ErrorWithAttrs) Error() string      { return e.Err.Error() }
func (e *ErrorWithAttrs) Unwrap() error      { return e.Err }
func (e *ErrorWithAttrs) Attrs() []slog.Attr { return e.attrs }

// Errorf creates a new error with optional slog.Attr attributes.
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
		Err:   fmt.Errorf(format, filteredArgs...),
		attrs: attrs,
	}
}

// AttrsFromError attempts to extract attributes from an error.
func AttrsFromError(err error) []slog.Attr {
	var ewa *ErrorWithAttrs
	if errors.As(err, &ewa) {
		return ewa.Attrs()
	}
	return nil
}

// Error converts an error into a slog.Group containing the error message and any attached attributes.
func Error(err error, groupKey string) slog.Attr {
	attrs := AttrsFromError(err)
	if len(attrs) == 0 {
		return slog.String("error", err.Error())
	}

	// Add the error message as an attribute inside the group
	attrs = append(attrs, slog.String("error", err.Error()))

	// Convert []slog.Attr to ...any for slog.Group
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}
	return slog.Group(groupKey, args...)
}
