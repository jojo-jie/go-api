package dayday

import (
	"database/sql"
	"runtime"
	"sync"
	"syscall"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"go.uber.org/goleak"
)

type File struct{ d int }

func openFile(path string) *File {
	d, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	p := &File{d}
	runtime.SetFinalizer(p, func(p *File) {
		syscall.Close(p.d)
	})

	return p
}

func readFile(descriptor int) string {
	doSomeAllocation()
	var buf [1000]byte
	_, err := syscall.Read(descriptor, buf[:])
	if err != nil {
		panic(err)
	}

	return string(buf[:])
}

func doSomeAllocation() {
	var a *int

	// memory increase to force the GC
	for i := 0; i < 10000000; i++ {
		i := 1
		a = &i
	}

	_ = a
}

func TestGC(t *testing.T) {
	//https://mp.weixin.qq.com/s/E3Lgl9T8iQYl65T3e0vDTA
	//快速找到 Goroutine 泄露的地方
	defer goleak.VerifyNone(t)
	//https://blog.csdn.net/qcrao/article/details/121571165
	p := openFile("one_test.go")
	content := readFile(p.d)
	//runtime.KeepAlive(p)
	println("Here is the content: " + content)
}

func TestSql(t *testing.T) {
	db, err := sql.Open("mysql", "root:xxxx@tcp(:)/my_test")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(20)
	defer db.Close()
	wg := sync.WaitGroup{}
	wg.Add(30)
	for i := 0; i < 30; i++ {
		go updateData(db, t, &wg)
	}
	wg.Wait()
}

func updateData(db *sql.DB, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	sqlStr := "update api_service set status=? where api_id = ? and status = 0"
	ret, err := db.Exec(sqlStr, 1, 8)
	if err != nil {
		t.Logf("update failed, err:%v\n", err)
		return
	}
	n, err := ret.RowsAffected() // 操作影响的行数
	if err != nil {
		t.Logf("get RowsAffected failed, err:%v\n", err)
		return
	}
	t.Logf("update success, affected rows:%d\n", n)
}
