package rank

import (
	"fmt"
	"testing"
)

func TestRankHeap(t *testing.T) {
	less := func(i, j int) bool { return i < j }
	sc := NewSortedHeap(5, less)

	for i := 1; i <= 100; i++ {
		sc.Insert(i)
	}

	fmt.Println(sc.Values()) // Prints: [5 4 3 2 1]
}
