package ringbuffer

import "errors"

var ErrIsEmpty = errors.New("ringbuffer is empty")

type RingBuffer[T any] struct {
	buf         []T
	initialSize int
	size        int
	r           int // read pointer
	w           int // write pointer
}

func NewRingBuffer[T any](initialSize int) *RingBuffer[T] {
	if initialSize <= 0 {
		panic("initial size must be great than zero")
	}

	// initial size must >= 2
	if initialSize == 1 {
		initialSize = 2
	}

	return &RingBuffer[T]{
		buf:         make([]T, initialSize),
		initialSize: initialSize,
		size:        initialSize,
	}
}

func (r *RingBuffer[T]) Read() (T, error) {
	var t T
	if r.r == r.w {
		return t, ErrIsEmpty
	}
	return nil,nil
}
