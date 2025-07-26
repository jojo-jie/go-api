package main

import (
	"archive/zip"
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

type emptyCtx int

func (*emptyCtx) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*emptyCtx) Done() <-chan struct{}       { return nil }
func (*emptyCtx) Err() error                  { return nil }
func (*emptyCtx) Value(key any) any           { return nil }

var background = new(emptyCtx)

func Background() Context {
	return background
}

type canceler interface {
	cancel(removeFromParent bool, err, cause error)
	Done() <-chan struct{}
}

type cancelCtx struct {
	Context

	mu       sync.Mutex
	done     atomic.Value
	children map[canceler]struct{}
	err      atomic.Value
	cause    error
}

func (c *cancelCtx) Done() <-chan struct{} {
	d := c.done.Load()
	if d != nil {
		return d.(chan struct{})
	}
	return nil
}

func (c *cancelCtx) Err() error {
	if e := c.err.Load(); e != nil {
		return e.(error)
	}
	return nil
}

func (c *cancelCtx) cancel(removeFromParent bool, err, cause error) {
	if err == nil {
		panic("err is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err.Load() != nil {
		return
	}

	c.err.Store(err)
	c.cause = cause

	if d, _ := c.done.Load().(chan struct{}); d != nil {
		close(d)
	}

	for child := range c.children {
		child.cancel(false, err, cause)
	}
	c.children = nil

	if removeFromParent {
		removeChild(c.Context, c)
	}
}

func WithCancel(parent Context) (Context, func(error)) {
	c := &cancelCtx{Context: parent}
	c.done.Store(make(chan struct{}))
	propagateCancel(parent, c)
	return c, func(e error) { c.cancel(true, e, nil) }
}

func propagateCancel(parent Context, child canceler) {
	if parent.Done() == nil {
		return
	}

	select {
	case <-parent.Done():
		child.cancel(false, parent.Err(), Cause(parent))
		return
	default:
	}

	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.err.Load() != nil {
			child.cancel(false, p.err.Load().(error), p.cause)
		} else {
			if p.children == nil {
				p.children = make(map[canceler]struct{})
			}
			p.children[child] = struct{}{}
		}
		return
	}

	go func() {
		select {
		case <-parent.Done():
			child.cancel(false, parent.Err(), Cause(parent))
		case <-child.Done():
		}
	}()
}

func parentCancelCtx(parent Context) (*cancelCtx, bool) {
	for {
		switch c := parent.(type) {
		case *cancelCtx:
			return c, true
		case interface{ Unwrap() Context }:
			parent = c.Unwrap()
		default:
			return nil, false
		}
	}
}

func removeChild(parent Context, child *cancelCtx) {
	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.children != nil {
			delete(p.children, child)
		}
	}
}

func Cause(ctx Context) error {
	if c, ok := ctx.(*cancelCtx); ok {
		return c.cause
	}
	return nil
}

func TestCCtx(t *testing.T) {
	t.Run("basic cancel", func(t *testing.T) {
		ctx, cancel := WithCancel(Background())
		ch := make(chan struct{})
		go func() {
			<-ctx.Done()
			close(ch)
		}()
		time.Sleep(10 * time.Millisecond)
		cancel(nil)
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for done")
		}
		if err := ctx.Err(); err != context.Canceled {
			t.Errorf("expected canceled, got %v", err)
		}
	})

	t.Run("propagate from parent", func(t *testing.T) {
		parent, pcancel := WithCancel(Background())
		ctx, _ := WithCancel(parent)
		pcancel(nil)
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for done")
		}
		if err := ctx.Err(); !reflect.DeepEqual(err, context.Canceled) {
			t.Errorf("expected canceled, got %v", err)
		}
	})
}

func ExampleCancelPropagation() {
	ctx, cancel := WithCancel(Background())

	go func() {
		<-ctx.Done()
		fmt.Println("child context canceled:", ctx.Err())
	}()

	time.Sleep(100 * time.Millisecond)
	cancel(nil)
	time.Sleep(100 * time.Millisecond)
	// Output:
	// child context canceled: context canceled
}

type WeekDay int

func (w WeekDay) Name() string {
	if w < Sunday || w > Monday {
		return "Unknown"
	}
	return [...]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}[w]
}

func (w WeekDay) Original() int {
	return int(w)
}

func (w WeekDay) String() string {
	return w.Name()
}

func Values() []WeekDay {
	return []WeekDay{Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday}
}

func ValueOf(name string) (WeekDay, error) {
	switch name {
	case "Sunday":
		return Sunday, nil
	case "Monday":
		return Monday, nil
	case "Tuesday":
		return Tuesday, nil
	case "Wednesday":
		return Wednesday, nil
	case "Thursday":
		return Thursday, nil
	case "Friday":
		return Friday, nil
	case "Saturday":
		return Saturday, nil
	default:
		return 0, fmt.Errorf("invalid WeekDay name: %s", name)
	}
}

const (
	Sunday WeekDay = iota + 1
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Monday
)

type Tree[E cmp.Ordered] struct {
	root *node[E]
}

func (t *Tree[E]) Insert(element E) {
	t.root = t.root.insert(element)
}

type node[E cmp.Ordered] struct {
	value E
	left  *node[E]
	right *node[E]
}

func (n *node[E]) insert(element E) *node[E] {
	if n == nil {
		return &node[E]{value: element}
	}
	switch {
	case element < n.value:
		n.left = n.left.insert(element)
	case element > n.value:
		n.right = n.right.insert(element)
	}
	return n
}

func TestE(t *testing.T) {

}

type FuncTree[E any] struct {
	root *funcNode[E]
	cmp  func(E, E) int
}

func NewFuncTree[E any](cmp func(E, E) int) *FuncTree[E] {
	return &FuncTree[E]{cmp: cmp}
}

func (t *FuncTree[E]) Insert(element E) {
	t.root = t.root.insert(t.cmp, element)
}

type funcNode[E any] struct {
	value E
	left  *funcNode[E]
	right *funcNode[E]
}

func (n *funcNode[E]) insert(cmp func(E, E) int, element E) *funcNode[E] {
	if n == nil {
		return &funcNode[E]{value: element}
	}
	sign := cmp(element, n.value)
	switch {
	case sign < 0:
		n.left = n.left.insert(cmp, element)
	case sign > 0:
		n.right = n.right.insert(cmp, element)
	}
	return n
}

type MethodTree[E Comparer[E]] struct {
	root *methodNode[E]
}

func (t *MethodTree[E]) Insert(element E) {
	t.root = t.root.insert(element)
}

type methodNode[E Comparer[E]] struct {
	value E
	left  *methodNode[E]
	right *methodNode[E]
}

func (n *methodNode[E]) insert(element E) *methodNode[E] {
	if n == nil {
		return &methodNode[E]{value: element}
	}
	sign := element.Compare(n.value)
	switch {
	case sign < 0:
		n.left = n.left.insert(element)
	case sign > 0:
		n.right = n.right.insert(element)
	}
	return n
}

type Comparer[T any] interface {
	Compare(T) int
}

func TestZip(t *testing.T) {
	file, err := os.OpenFile("my_contents.zip", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open file", "error", err)
	}

	writer := zip.NewWriter(file)
	err = writer.AddFS(os.DirFS("/home/denis/Pictures/not_in_train/zip_test"))
	if err != nil {
		slog.Error("failed to write files to zip archive", "error", err)
	}

	err = writer.Close()
	if err != nil {
		slog.Error("failed to close zip writer", "error", err)
	}
}

type Queue []int

// Enqueue adds an element to the rear of the queue
func (q *Queue) Enqueue(value int) {
	*q = append(*q, value)
}

// Dequeue removes and returns an element from the front of the queue
func (q *Queue) Dequeue() (int, error) {
	if q.IsEmpty() {
		return 0, fmt.Errorf("empty queue")
	}
	value := (*q)[0]
	(*q)[0] = 0 // Zero out the element (optional)
	*q = (*q)[1:]
	return value, nil
}

// CheckFront returns the front element without removing it
func (q *Queue) CheckFront() (int, error) {
	if q.IsEmpty() {
		return 0, fmt.Errorf("empty queue")
	}
	return (*q)[0], nil
}

// IsEmpty checks if the queue is empty
func (q *Queue) IsEmpty() bool {
	return len(*q) == 0
}

// Size returns the number of elements in the queue
func (q *Queue) Size() int {
	return len(*q)
}

// PrintQueue displays all elements in the queue
func (q *Queue) PrintQueue() {
	if q.IsEmpty() {
		fmt.Println("Queue is empty")
		return
	}
	for _, item := range *q {
		fmt.Printf("%d ", item)
	}
	fmt.Println()
}
