package dayday

import (
	"dayday/constraints"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// https://mp.weixin.qq.com/s?__biz=MzUxMDI4MDc1NA==&mid=2247493705&idx=1&sn=3d6ca99f97086b13e3586f03ecad0bf2&scene=21#wechat_redirect
// 泛型 类型参数 类型约束 类型推导
// add slice
func TestAddSlice(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := reduce[int](numbers, func(acc, current int) int {
		return acc + current
	}, 0)
	t.Log(sum)

	divided := reduce[int, float64](numbers, func(acc float64, current int) float64 {
		return acc + float64(current)/10.0
	}, 0)
	t.Log(divided)
}

func reduce[T, M any](s []T, f func(M, T) M, initValue M) M {
	acc := initValue
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func cmp[T Ordered](p0, p1 T) bool {
	return p0 < p1
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func TestSliceMap(t *testing.T) {
	numbers := []float64{4, 9, 16, 25}
	t.Log(mapSlice[float64, float64](numbers, math.Sqrt))

	words := []string{"a", "b", "c", "d"}
	t.Log(mapSlice[string, string](words, func(s string) string {
		return "\"" + strings.ToUpper(s) + "\""
	}))

	stringPowNumbers := mapSlice(numbers, func(n float64) string {
		return strconv.FormatFloat(math.Pow(n, 2), 'f', -1, 64)
	})
	t.Log(stringPowNumbers)
}

func mapSlice[T, M any](a []T, f func(T) M) []M {
	n := make([]M, len(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

func TestSliceFilter(t *testing.T) {
	websites := []string{"https://www.baidu.com", "https://baidu.com", "http://pay.baidu.com", "http://zhidao.baidu.com"}
	t.Log(filter(websites, func(w string) bool {
		return strings.HasPrefix(w, "https")
	}))

	numbers := []int{1, 2, 3, 4, 5, 6}
	t.Log(filter(numbers, func(n int) bool {
		return n%2 == 0
	}))
}

func filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

func TestSortSlice(t *testing.T) {
	floatSlice := []float64{2.3, 1.2, 0.2, 51.2}
	sortSlice[float64](floatSlice)
	t.Log(floatSlice)

	stringSlice := []string{"z", "a", "b"}
	sortSlice(stringSlice)
	t.Log(stringSlice)

	intSlice := []int{0, 3, 2, 1, 6}
	sortSlice(intSlice)
	t.Log(intSlice)
}

func sortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func TestMapKeys(t *testing.T) {
	vegetableSet := map[string]bool{
		"potato":  true,
		"cabbage": true,
		"carrot":  true,
	}
	t.Log(keys(vegetableSet))

	fruitRank := map[int]string{
		1: "strawberry",
		2: "raspberry",
		3: "blueberry",
	}
	t.Log(keys(fruitRank))
}

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestContains(t *testing.T) {
	t.Log(contains([]string{"a", "b", "c"}, "b"))
	t.Log(contains([]int{1, 2, 3}, 2))
	t.Log(contains([]int{1, 2, 3}, 10))
}

func contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if s == v {
			return true
		}
	}
	return false
}

func TestMinMax(t *testing.T) {
	t.Log(max([]int{10, 2, 4, 1, 6, 8, 2}, "min"))
}

func max[T constraints.Ordered](s []T, t string) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	m := s[0]
	for _, v := range s {
		if "max" == t && m < v {
			m = v
		}
		if "min" == t && m > v {
			m = v
		}
	}
	return m
}
