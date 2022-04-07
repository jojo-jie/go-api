package dayday

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

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

type Q interface {
	~int | ~int8 | ~int32 | ~int64 | ~float64 | ~float32
}

//nlogn
func TestQuickSort(t *testing.T) {
	rand.Seed(time.Now().Unix())
	sequence := rand.Perm(33)
	t.Logf("sequence before sort: %v", sequence)
	quickSort[int](sequence, 0, len(sequence)-1)
	t.Logf("sequence after sort: %v", sequence)
}

func quickSort[T Q](sequence []T, low, high int) {
	if high <= low {
		return
	}
	i := partition[T](sequence, low, high)
	quickSort[T](sequence, low, i-1)
	quickSort[T](sequence, i+1, high)
}

func partition[T Q](sequence []T, low, high int) int {
	i, j := low, high-1
	for {
		for sequence[i] < sequence[high] {
			i++
			if i >= high {
				break
			}
		}

		for sequence[j] > sequence[high] {
			j--
			if j <= low {
				break
			}
		}

		if i >= j {
			break
		}
		sequence[i], sequence[j] = sequence[j], sequence[i]
	}
	sequence[high], sequence[i] = sequence[i], sequence[high]
	return i
}

type sumT interface {
	~int | ~float64 | ~float32
}

func TestGeneric(t *testing.T) {
	rand.Seed(time.Now().Unix())
	list := rand.Perm(10)
	t.Logf("list %v", list)
	t.Logf("sum %v", sumSlice[int](list))
}

func sumSlice[T sumT](s []T) T {
	var ss T
	for _, v := range s {
		ss = ss + v
	}
	return ss
}

type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%s: %d", p.Name, p.Age)
}

type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].Age > a[j].Age }

func TestSort(t *testing.T) {
	p := []Person{
		{"kirito", 17},
		{"asuna", 19},
		{"baobao", 22},
		{"linjie", 19},
	}
	t.Logf("sort before %+v", p)
	sort.Sort(ByAge(p))
	t.Logf("sort after %+v", p)
}

//nlogn
func TestSortMerge(t *testing.T) {
	rand.Seed(time.Now().Unix())
	sequence := rand.Perm(33)
	t.Logf("sequence before sort: %v", sequence)
	MergeSort[int](sequence, 0, len(sequence))
	t.Logf("sequence after sort: %v", sequence)
}

func MergeSort[T Q](array []T, begin int, end int) {
	if end-begin > 1 {
		mid := begin + (end-begin+1)/2
		MergeSort[T](array, begin, mid)
		MergeSort[T](array, mid, end)
		merge(array, begin, mid, end)
	}
}

func merge[T Q](array []T, begin, mid, end int) {
	leftSize := mid - begin
	rightSize := end - mid
	newSize := leftSize + rightSize
	result := make([]T, 0, newSize)
	l, r := 0, 0
	for l < leftSize && r < rightSize {
		lValue := array[begin+l]
		rValue := array[mid+r]
		if lValue < rValue {
			result = append(result, lValue)
			l++
		} else {
			result = append(result, rValue)
			r++
		}
	}

	result = append(result, array[begin+l:mid]...)
	result = append(result, array[mid+r:end]...)

	for i := 0; i < newSize; i++ {
		array[begin+i] = result[i]
	}
	return
}
