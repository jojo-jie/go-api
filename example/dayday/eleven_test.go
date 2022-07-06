package dayday

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/lucas-clemente/quic-go/http3"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestCache(t *testing.T) {
	gc := gcache.New(20).
		LRU().
		Build()
	gc.Set("key", "ok")
	value, err := gc.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Get:", value)
}

func TestHttp3(t *testing.T) {
	certFile := ""
	keyFile := ""
	handle := func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "hello world")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handle)
	err := http3.ListenAndServe(":443", certFile, keyFile, mux)
	fmt.Println(err)
}

func TestReusable(t *testing.T) {
	n, err := transfer()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	t.Logf("%d bytes transferred.\n", n)
}

func transfer() (n int64, err error) {
	if n, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		return n, err
	}
	return n, err
}

func ioCopy(dst, src *os.File) error {
	buf := make([]byte, 32768)
	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			_, ew := dst.Write(buf[:nr])
			if ew != nil {
				return ew
			}
		}

		if io.EOF == err {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// write example.
func write(dst *os.File, buf []byte) error {
	nw, ew := dst.Write(buf)

	fmt.Printf("wrote    : %d bytes\n", nw)
	fmt.Printf("write err: %v\n", ew)

	return ew
}

// read example.
func read(src *os.File) error {
	buf := make([]byte, 1024*32) // in the middle.
	// buf := make([]byte, 148157) // defies the purpose of streaming.
	// buf := make([]byte, 8) // too many chunking.

	for {
		nr, er := src.Read(buf)
		// fmt.Printf("buf      : %q\n", buf[0:nr])
		fmt.Printf("read     : %d bytes\n", nr)
		fmt.Printf("read err : %v\n", er)

		if er == io.EOF {
			return nil
		}
		if er != nil {
			return er
		}
	}
}

func TestPointer(t *testing.T) {
	t.Log("••••• ARRAYS")
	arrays()

	t.Log("\n••••• SLICES")
	slices()

	t.Log("\n••••• MAPS")
	maps()

	t.Log("\n••••• STRUCTS")
	structs()
}

type house struct {
	name  string
	rooms int
}

func structs() {
	myHouse := house{name: "My House", rooms: 5}

	addRoom(myHouse)

	// fmt.Printf("%+v\n", myHouse)
	fmt.Printf("structs()     : %p %+v\n", &myHouse, myHouse)

	addRoomPtr(&myHouse)
	fmt.Printf("structs()     : %p %+v\n", &myHouse, myHouse)
}

func addRoomPtr(h *house) {
	h.rooms++ // same: (*h).rooms++
	fmt.Printf("addRoomPtr()  : %p %+v\n", h, h)
	fmt.Printf("&h.name       : %p\n", &h.name)
	fmt.Printf("&h.rooms      : %p\n", &h.rooms)
}

func addRoom(h house) {
	h.rooms++
	fmt.Printf("addRoom()     : %p %+v\n", &h, h)
}

// ••••••••••••••••••••••••••••••••••••••••••••••••••

func maps() {
	confused := map[string]int{"one": 2, "two": 1}
	fix(confused)
	fmt.Println(confused)

	// &confused["one"]
}

func fix(m map[string]int) {
	m["one"] = 1
	m["two"] = 2
	m["three"] = 3
}

// ••••••••••••••••••••••••••••••••••••••••••••••••••

func slices() {
	dirs := []string{"up", "down", "left", "right"}

	up(dirs)
	fmt.Printf("slices list   : %p %q\n", &dirs, dirs)

	upPtr(&dirs)
	fmt.Printf("slices list   : %p %q\n", &dirs, dirs)
}

func upPtr(list *[]string) {
	lv := *list

	for i := range lv {
		lv[i] = strings.ToUpper(lv[i])
	}

	*list = append(*list, "HEISEN BUG")

	fmt.Printf("upPtr list    : %p %q\n", list, list)
}

func up(list []string) {
	for i := range list {
		list[i] = strings.ToUpper(list[i])
		fmt.Printf("up.list[%d]    : %p\n", i, &list[i])
	}

	list = append(list, "HEISEN BUG")

	fmt.Printf("up list       : %p %q\n", &list, list)
}

// ••••••••••••••••••••••••••••••••••••••••••••••••••

func arrays() {
	nums := [...]int{1, 2, 3}

	incr(nums)
	fmt.Printf("arrays nums   : %p\n", &nums)
	fmt.Println(nums)

	incrByPtr(&nums)
	fmt.Println(nums)
}

func incr(nums [3]int) {
	fmt.Printf("incr nums     : %p\n", &nums)
	for i := range nums {
		nums[i]++
		fmt.Printf("incr.nums[%d]  : %p\n", i, &nums[i])
	}
}

func incrByPtr(nums *[3]int) {
	fmt.Printf("incrByPtr nums: %p\n", &nums)
	for i := range nums {
		nums[i]++ // same: (*nums)[i]++
	}
}
