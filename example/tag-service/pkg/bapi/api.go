package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"io"
	"net/http"
	"tag-service/pkg/tracer"
	"time"
)

const (
	APP_KEY    = "eddycjy"
	APP_SECRET = "go-programming-tour-book"
)

type AccessToken struct {
	Token string `json:"token"`
}

func (a *API) getAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf(
		"%s?app_key=%s&app_secret=%s",
		"auth",
		APP_KEY,
		APP_SECRET,
	)
	body, err := a.httpGet(ctx, url)
	if err != nil {
		return "", err
	}
	var accessToken AccessToken
	err = json.Unmarshal(body, &accessToken)
	if err != nil {
		return "", err
	}
	return accessToken.Token, nil
}

func (a *API) httpGet(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", a.URL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	//newCtx, finish:=tracer.Start(opentracing.GlobalTracer(), "HTTP GET: "+a.URL, ctx, req)
	/*span, newCtx := opentracing.StartSpanFromContext(ctx, "HTTP GET: "+a.URL, opentracing.Tag{
		Key:   string(ext.Component),
		Value: "HTTP",
	})
	span.SetTag("url", url)*/
	newCtx, finish := tracer.Start(opentracing.GlobalTracer(), "HTTP GET: "+a.URL, ctx, req)

	//在jaeger-client-go库中也是通过类似的操作去传递信息
	//Inject 和 Extract 操作
	//在单体程序中, 父子Span通过Span Context关联, 而Span Context是在内存中的, 显而易见这样的方法在跨应用的场景下是行不通的。
	//
	//跨应用跨进程通讯使用的方式通常是"序列化"，在jaeger-client-python库中也是通过类似的操作去传递信息, 它们叫:Tracer.inject() 与 Tracer.extract()。
	//
	//当客户端发起http通信时候，当前进程调用Tracer.inject(…)注入当前活动的Span Context以及其他相关参数，通常客户端可以将该Span Context以 http 的 headers 参数(trace_id)的方式标识传递， 服务进程调用Tracer.extract(…)，从传入的请求的headers中抽取从上面注入的Span Context和参数还原上下文。
	//原文链接：https://blog.csdn.net/pushiqiang/article/details/114449564
	// = opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	finish(tracer.InjectHttp(), tracer.SetTag())
	req = req.WithContext(newCtx)
	client := http.Client{Timeout: time.Second * 3}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type API struct {
	URL string
}

func NewAPI(url string) *API {
	return &API{
		URL: url,
	}
}

func (a API) GetTagList(ctx context.Context, name string) ([]byte, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := a.httpGet(ctx, fmt.Sprintf("%s?token=%s&name=%s", "api/v1/tags", token, name))
	if err != nil {
		return nil, err
	}
	return body, nil
}
