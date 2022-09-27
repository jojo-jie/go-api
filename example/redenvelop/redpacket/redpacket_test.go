package redpacket

import (
	"testing"
)

func TestNewRedPacketPool(t *testing.T) {
	p, _ := NewRedPacketPool(float64(10), 10)
	for _, v := range p {
		t.Logf("%+v", *v)
	}
}
