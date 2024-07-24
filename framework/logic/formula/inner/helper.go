// All rights reserved. This is part of West Securities ltd. proprietary source code.
// No part of this file may be reproduced or transmitted in any form or by any means,
// electronic or mechanical, including photocopying, recording, or by any information
// storage and retrieval system, without the prior written permission of West Securities ltd.

// author:  Maminghui (Digital Office Product Department #2)

package indicator

import (
	"math"
	"testing"
)

// Test equals.
func testEquals(t *testing.T, actual, expected []float64) {
	if len(actual) != len(expected) {
		t.Fatal("not the same size")
	}

	for i := 0; i < len(expected); i++ {
		if actual[i] != expected[i] {
			t.Fatalf("at %d actual %f expected %f", i, actual[i], expected[i])
		}
	}
}

// Check values same size.
func checkSameSize(values ...[]float64) {
	if len(values) < 2 {
		return
	}

	n := len(values[0])

	for i := 1; i < len(values); i++ {
		if len(values[i]) != n {
			panic("not all same size")
		}
	}
}

// Round value to digits.
func roundDigits(value float64, digits int) float64 {
	n := math.Pow(10, float64(digits))

	return math.Round(value*n) / n
}

// Round values to digits.
func roundDigitsAll(values []float64, digits int) []float64 {
	result := make([]float64, len(values))

	for i := 0; i < len(result); i++ {
		result[i] = roundDigits(values[i], digits)
	}

	return result
}

// Multiply values by multipler.
func multiplyBy(values []float64, multiplier float64) []float64 {
	result := make([]float64, len(values))

	for i, value := range values {
		result[i] = value * multiplier
	}

	return result
}

// Add values1 and values2.
func add(values1, values2 []float64) []float64 {
	checkSameSize(values1, values2)

	result := make([]float64, len(values1))
	for i := 0; i < len(result); i++ {
		result[i] = values1[i] + values2[i]
	}

	return result
}

// subtract values2 from values1.
func subtract(values1, values2 []float64) []float64 {
	subtract := multiplyBy(values2, float64(-1))
	return add(values1, subtract)
}

// Divide values by divider.
func divideBy(values []float64, divider float64) []float64 {
	return multiplyBy(values, float64(1)/divider)
}

// Divide values1 by values2.
func divide(values1, values2 []float64) []float64 {
	checkSameSize(values1, values2)

	result := make([]float64, len(values1))

	for i := 0; i < len(result); i++ {
		result[i] = values1[i] / values2[i]
	}

	return result
}
