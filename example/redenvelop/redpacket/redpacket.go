package redpacket

import (
	"errors"
	"math/rand"
	"snowflake"
	"strconv"
	"time"
)

// RedPacket 红包
type RedPacket struct {
	RedPacketId string  `json:"redPacketId"`
	Amount      float64 `json:"amount"`
}

// Pool 红包池
type Pool []*RedPacket

var w *snowflake.Worker

func init() {
	w = snowflake.NewWorker(5, 5)
}

// NewRedPacket 创建单个红包信息 单位分
func NewRedPacket(amount float64) (*RedPacket, error) {
	id, err := w.NextID()
	if err != nil {
		return nil, err
	}
	return &RedPacket{
		RedPacketId: strconv.Itoa(int(id)),
		Amount:      amount,
	}, err
}

func NewRedPacketPool(amount float64, n int) (Pool, error) {
	l, err := redAmount2N(amount, n)
	if err != nil {
		return nil, err
	}
	rp := make(Pool, 0, n)
	for _, a := range l {
		if r, err := NewRedPacket(float64(a) / 100); err != nil {
			return nil, err
		} else {
			rp = append(rp, r)
		}
	}
	return rp, nil
}

// 随机红包额度 金额分
func redAmount(amount float64, n int) ([]int, error) {
	fen := int(amount * 100)
	if fen < n || fen < 1 {
		return nil, errors.New("被拆分的总金额不能小于1分")
	}
	s := make([]int, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, 1)
	}
	fen -= n
	var i int
	for fen > 1 {
		rand.Seed(time.Now().UnixNano())
		f := rand.Intn(fen)
		i++
		s[i%n] += f
		fen = fen - f
	}
	if fen > 0 {
		s[0] += fen
	}
	return s, nil
}

func redAmount2N(amount float64, n int) ([]int, error) {
	fen := int(amount * 100)
	if fen < n || fen < 1 {
		return nil, errors.New("被拆分的总金额不能小于1分")
	}
	s := make([]int, n, n)
	remain := fen
	sum := 0
	for i := 0; i < n; i++ {
		x := doubleAverage(n-i, remain)
		remain -= x
		sum += x
		s[i] = x
	}
	return s, nil
}

var min int = 1

func doubleAverage(count, amount int) int {
	if count == 1 {
		return amount
	}
	max := amount - min*count
	avg := max / count
	avg2 := 2*avg + min
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(avg2) + min
	return x
}
