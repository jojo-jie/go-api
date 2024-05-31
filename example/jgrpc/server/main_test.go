package main

import (
	"reflect"
	"testing"
)

type MyType struct {
	Value int
}

type MyType2 struct {
	Value int
}

/*// 实现 == 操作符
func (a MyType) Equal(b MyType) bool {
	return a.Value == b.Value
}

// 实现 != 操作符
func (a MyType) NotEqual(b MyType) bool {
	return a.Value != b.Value
}*/

func TestComp(t *testing.T) {
	a := MyType{Value: 1}
	b := MyType2{Value: 1}

	t.Log(reflect.ValueOf(a).Equal(reflect.ValueOf(b)))
	v := reflect.ValueOf(a)
	t.Log(v.Type())
	t.Log(compare(a, b))
}

func compare(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
