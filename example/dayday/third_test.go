package dayday

import (
	"dayday/base_demo"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//https://mp.weixin.qq.com/s/etqhRm0ci_xhLovb2RhNJw
//go 单元测试
// 表格驱动测试
type tableStruct struct {
	name  string
	input string
	sep   string
	want  []string
}

var tests []tableStruct

func init() {
	tests = []tableStruct{
		{"base case", "a:b:c", ":", []string{"a", "b", "c"}},
		{"wrong sep", "a:b:c", ",", []string{"a:b:c"}},
		{"more sep", "abcd", "bc", []string{"a", "d"}},
		{"leading sep", "沙河有沙又有河", "沙", []string{"", "河有", "又有河"}},
	}
}

func TestSplitAll(t *testing.T) {
	// 断言
	assertions := assert.New(t)
	for _, tt := range tests {
		// 使用t.Run()执行子测试
		t.Run(tt.name, func(t *testing.T) {
			got := base_demo.Split(tt.input, tt.sep)
			// slice 无法直接比较
			assertions.Equal(tt.want, got, "they should be equal")
		})
	}
}

func TestSplitAllConcurrent(t *testing.T) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() //将每个测试用例标记为能够彼此并行运行
			got := base_demo.Split(tt.input, tt.sep)
			require.Equal(t, tt.want, got)
		})
	}
}

//https://mp.weixin.qq.com/s/IBZO0jakAAeHgl8OBk21IA
//模拟服务请求和接口返回
func TestHelloHandler(t *testing.T) {
	tests := []struct {
		name   string
		param  string
		expect string
	}{
		{"base case", `{"name": "liwenzhou"}`, "hello liwenzhou"},
		{"bad case", "", "we need a name"},
	}
	r := base_demo.SetupRouter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// mock one http request
			req := httptest.NewRequest(http.MethodPost, "/hello", strings.NewReader(tt.param))
			// mock one http response
			w := httptest.NewRecorder()
			// 让server端处理mock请求并记录返回的响应内容
			r.ServeHTTP(w, req)
			// 校验状态码是否符合预期
			assert.Equal(t, http.StatusOK, w.Code)

			var resp map[string]string
			err := json.Unmarshal([]byte(w.Body.String()), &resp)
			assert.Nil(t, err)
			assert.Equal(t, tt.expect, resp["msg"])
		})
	}
}
