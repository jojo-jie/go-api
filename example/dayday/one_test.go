package dayday

import (
	"math/rand"
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

func TestQuickSort(t *testing.T) {
	rand.Seed(time.Now().Unix())
	sequence := rand.Perm(34)
	t.Logf("sequence before sort: %v", sequence)
	quickSort(sequence, 0, len(sequence)-1)
	t.Logf("sequence after sort: %v", sequence)
}

func quickSort(sequence []int, low int, high int) {
	if high <= low {
		return
	}
	j := partition(sequence, low, high)
	quickSort(sequence, low, j-1)
	quickSort(sequence, j+1, high)
}

// first quicksort
func partition(sequence []int, low, high int) int {
	i, j := low+1, high
	for {
		// first element is pivot
		for sequence[i] < sequence[low] {
			i++
			if i >= high {
				break
			}
		}

		for sequence[j] > sequence[low] {
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
	sequence[low], sequence[j] = sequence[j], sequence[low]
	return j
}
