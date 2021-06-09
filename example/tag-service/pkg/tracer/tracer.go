package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"time"
)

// NewJaegerTracer 创建JaegerTracer 对象及基本配置
func NewJaegerTracer(serviceName, agentHostPort string) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		//服务名
		ServiceName: serviceName,
		//取样器，采样模式 const 固定采样
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentHostPort,
		},
	}
	// 初始化tracer 对象 opentracing.Tracer 并不是某个供应商的tracer 对象
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}

	// 设置全局tracer 对象
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer, nil
}
