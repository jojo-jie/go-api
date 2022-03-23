package dayday

import "testing"

// https://mp.weixin.qq.com/s/0nChJVXxhFYDu453e7Z8Hg
func TestCC(t *testing.T) {
	done, _ := ccc()
	done()
}

func aaa() (done func(), err error) {
	return func() {
		print("aaa:done")
	}, nil
}

// 陷阱
// return 实际是个复制语句，最后执行 return 语句后，会对返回变量 done 进行赋值
func bbb() (done func(), err error) {
	done, err = aaa()
	return func() {
		print("bbb:surprise!")
		done()
	}, nil
}

func ccc() (func(), error) {
	done, _ := aaa()
	return func() {
		print("ccc:surprise!")
		done()
	}, nil
}
