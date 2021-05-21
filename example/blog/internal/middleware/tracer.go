package middleware

import (
	"blog/global"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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
			span, newCtx = opentracing.StartSpanFromContextWithTracer(c.Request.Context(), global.Tracer, c.Request.URL.Path)
		} else {
			span, newCtx = opentracing.StartSpanFromContextWithTracer(c.Request.Context(), global.Tracer, c.Request.URL.Path, opentracing.ChildOf(spanCtx), opentracing.Tag{Key: string(ext.Component), Value: "HTTP"})
		}
		defer span.Finish()
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}
