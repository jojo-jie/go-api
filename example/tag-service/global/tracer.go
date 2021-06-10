package global

import (
	"github.com/opentracing/opentracing-go"
	"io"
)

var (
	Tracer opentracing.Tracer
	Closer io.Closer
)
