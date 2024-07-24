package strategy

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test arithmeticSequence
func TestArithmeticSequence(t *testing.T) {
	total := 1000.0
	n := 10
	ratio := 0.5
	amt := arithmeticSequence(total, n, ratio)
	if len(amt) != n {
		t.Errorf("len(amt) = %d, want %d", len(amt), n)
	}
	// sum of the arithmetic sequence
	sum := 0.0
	for _, val := range amt {
		sum += val
	}
	if math.Abs(sum-total) > 1e-6 {
		t.Errorf("sum = %f, want %f", sum, total)
	}
}


// test rank.NewSortedHeap

func TestRankHeap(t *testing.T) {

	tmp := []float64{6.0, 7.0, 8.0, 9.0, 10.0, 1.0, 2.0, 3.0, 4.0, 5.0 }

	sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j]})
	// insert 10 signals, only keep last 5
	
	fmt.Println(tmp)
	// slice should be [1 2 3 4 5 6 7 8 9 10]
	expected := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	assert.Equal(t, tmp, expected)

}

// test filterSignal
func TestFilterSignal(t *testing.T) {
	signals := []SortRank{
		SortRank{instID: "1", signal: 1, close: 1.0},
		SortRank{instID: "2", signal: -2, close: 2.0},
		SortRank{instID: "3", signal: 3, close: 3.0},
		SortRank{instID: "4", signal: -4, close: 4.0},
		SortRank{instID: "5", signal: 5, close: 5.0},
	}
	// filter out signals with signal > 3
	filtered := filterSignal(signals, func(signal SortRank) bool {
		return signal.signal > 0.0
	})
	// only 4 and 5 are left
	expected := []SortRank{
		SortRank{instID: "1", signal: 1, close: 1.0},
		SortRank{instID: "3", signal: 3, close: 3.0},
		SortRank{instID: "5", signal: 5, close: 5.0},
	}
	assert.Equal(t, filtered, expected)
}