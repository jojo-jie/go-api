package ppp

import (
	"fmt"
	"slices"
	"testing"
)

func TestConcat(t *testing.T) {
	s1 := []string{"Go slices", "Go maps"}
	s2 := []string{"Go strings", "Go strconv"}
	s3 := []string{"joker", "jojo"}
	s4 := slices.Concat(s1, s2, s3)
	fmt.Printf("cap: %d, len: %d\n", cap(s4), len(s4))
	fmt.Println(s4)
}

func TestReplace(t *testing.T) {
	s1 := []int{1, 6, 7, 4, 5}
	s2 := slices.Replace(s1, 1, 4, 2)
	fmt.Println(s1)
	fmt.Println(s2)
}
