package dayday

import "testing"

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
