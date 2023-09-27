package main

import (
	"bytes"
	"math/rand"
	_ "net/http/pprof"
	"time"
)

func genRandomBytes() *bytes.Buffer {
	var buf bytes.Buffer
	for i := 0; i < 10000; i++ {
		buf.Write([]byte{'0' + byte(rand.Intn(10))})
	}
	return &buf
}

const shortDuration = 1 * time.Millisecond

func main() {
	/*go func() {
		for i := 0; i < 10000; i++ {
			_ = genRandomBytes()
		}
		time.Sleep(time.Second)
	}()*/

	//log.Fatal(http.ListenAndServe(":6060", nil))

	/*tooSlow := fmt.Errorf("too slow")
	ctx, cancel := context.WithTimeoutCause(context.Background(), shortDuration, tooSlow)
	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(context.Cause(ctx))
	}*/

	twoCtxCancel()
}

func twoCtxCancel() {
	/*finishedEarly := fmt.Errorf("finished early")
	tooSlow := fmt.Errorf("too slow")
	ctx, _ := context.WithTimeoutCause(context.Background(), 1*time.Second, tooSlow)
	ctx, cancel := context.WithCancelCause(ctx)
	time.Sleep(2 * time.Second)
	stopf := context.AfterFunc(ctx, func() {
		fmt.Println("stopf....")
		cancel(finishedEarly)
		fmt.Println(context.Cause(ctx))
		fmt.Println(ctx.Err())
	})
	fmt.Println(stopf())
	<-time.After(1 * time.Second)*/
	/*slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	slog.Info("hello", "name", "Al")
	slog.Error("oops", net.ErrClosed, "status", 500)
	slog.LogAttrs(context.Background(), slog.LevelDebug, "sss",
		slog.Int("status", 500), slog.Any("err", net.ErrClosed))*/
}

type Tries struct {
	child [26]*Tries
	isEnd bool
}

func (t *Tries) Insert(word string) {
	cur := t
	for i := 0; i < len(word); i++ {
		idx := word[i] - 'a'
		if cur.child[idx] == nil {
			cur.child[idx] = &Tries{}
		}
		cur = cur.child[idx]
	}
	cur.isEnd = true
}

func (t *Tries) search(word string) bool {
	cur := t
	for i := 0; i < len(word); i++ {
		if cur.child[word[i]-'a'] == nil {
			return false
		}
		cur = cur.child[word[i]-'a']
	}
	return cur.isEnd
}
