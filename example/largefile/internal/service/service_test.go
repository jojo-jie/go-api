package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

// 定义结构体
type Part struct {
	PartNumber     int    `json:"PartNumber"`
	ETag           string `json:"ETag"`
	ChecksumCRC32  string `json:"ChecksumCRC32"`
	ChecksumCRC32C string `json:"ChecksumCRC32C"`
	ChecksumSHA1   string `json:"ChecksumSHA1"`
	ChecksumSHA256 string `json:"ChecksumSHA256"`
}

// 定义二维切片
type PartsSlice [][]Part

// 实现Less函数，用于比较两个Part对象的PartNumber字段
func (p PartsSlice) Less(i, j int) bool {
	return p[i][0].PartNumber < p[j][0].PartNumber
}

// 实现Len函数，返回二维切片中元素的数量
func (p PartsSlice) Len() int {
	return len(p)
}

// 实现Swap函数，用于交换两个Part对象的位置
func (p PartsSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func TestSortD(t *testing.T) {
	// 创建一个二维切片
	parts := PartsSlice{
		[]Part{{4, "0123456789abcdef0123456789abcdef", "", "", "", ""},
			{2, "fedcba9876543210fedcba9876543210", "", "", "", ""}},
		[]Part{{11, "c4403d54a23fecc999b3c6b4487b25e8", "", "", "", ""},
			{6, "76db3bb804f6ef0c6335ec8493db1946", "", "", "", ""}},
	}

	// 对二维切片中的每个一维切片进行排序
	for i := range parts {
		sort.Slice(parts[i], func(j, k int) bool {
			return parts[i][j].PartNumber < parts[i][k].PartNumber
		})
	}

	// 对整个二维切片按照第一个元素排序
	sort.Sort(parts)
	p, _ := json.Marshal(parts)
	t.Logf("Sorted Parts:\n %s", string(p))

	for i := range 10 {
		t.Log(i)
	}
}

var resource int32 = 0

func Read(rwm *sync.RWMutex, wg *sync.WaitGroup) {

	rwm.RLock()
	defer rwm.RUnlock()
	fmt.Println("read lock acquired")

	time.Sleep(time.Second * 3)

	fmt.Println("read lock released")
	wg.Done()
}

func Write(rwm *sync.RWMutex, wg *sync.WaitGroup) {

	rwm.Lock()
	defer rwm.Unlock()
	fmt.Println("write lock acquired")

	resource = resource + 1
	time.Sleep(time.Second * 3)

	fmt.Println("write lock released")
	wg.Done()

}

func TestRWM(t *testing.T) {
	rwm := &sync.RWMutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(2)
		go Write(rwm, wg)
		go Read(rwm, wg)
	}
	wg.Wait()
}
