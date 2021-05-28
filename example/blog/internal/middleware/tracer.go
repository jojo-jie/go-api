package middleware

import (
	"blog/global"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
)

func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		var newCtx context.Context
		var span opentracing.Span
		spanCtx, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)
		if err != nil {
			span, newCtx = opentracing.StartSpanFromContextWithTracer(c.Request.Context(), global.Tracer, c.Request.URL.Path, opentracing.Tag{Key: string(ext.Component), Value: "HTTP"})
		} else {
			span, newCtx = opentracing.StartSpanFromContextWithTracer(c.Request.Context(), global.Tracer, c.Request.URL.Path, opentracing.ChildOf(spanCtx), opentracing.Tag{Key: string(ext.Component), Value: "HTTP"})
		}
		defer span.Finish()
		var traceID string
		var SpanID string
		var spanContext = span.Context()
		switch spanContext.(type) {
		case jaeger.SpanContext:
			traceID = spanContext.(jaeger.SpanContext).TraceID().String()
			SpanID = spanContext.(jaeger.SpanContext).SpanID().String()
			c.Set("X-Trace-ID", traceID)
			c.Set("X-Span-ID", SpanID)
			c.Request = c.Request.WithContext(newCtx)
			c.Next()
		}
	}
}
