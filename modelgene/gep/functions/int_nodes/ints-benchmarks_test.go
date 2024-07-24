// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Code generated by gen-benchmarks; DO NOT EDIT.

package intNodes

import (
	"testing"
)

func BenchmarkMultiply(b *testing.B) { runBenchmark(b, "*") }
func BenchmarkPlus(b *testing.B)     { runBenchmark(b, "+") }
func BenchmarkMinus(b *testing.B)    { runBenchmark(b, "-") }
func BenchmarkDivide(b *testing.B)   { runBenchmark(b, "/") }
func BenchmarkAdd3(b *testing.B)     { runBenchmark(b, "Add3") }
func BenchmarkAdd4(b *testing.B)     { runBenchmark(b, "Add4") }
func BenchmarkAvg2(b *testing.B)     { runBenchmark(b, "Avg2") }
func BenchmarkAvg3(b *testing.B)     { runBenchmark(b, "Avg3") }
func BenchmarkAvg4(b *testing.B)     { runBenchmark(b, "Avg4") }
func BenchmarkDiv3(b *testing.B)     { runBenchmark(b, "Div3") }
func BenchmarkDiv4(b *testing.B)     { runBenchmark(b, "Div4") }
func BenchmarkET2A(b *testing.B)     { runBenchmark(b, "ET2A") }
func BenchmarkET2B(b *testing.B)     { runBenchmark(b, "ET2B") }
func BenchmarkET2C(b *testing.B)     { runBenchmark(b, "ET2C") }
func BenchmarkET2E(b *testing.B)     { runBenchmark(b, "ET2E") }
func BenchmarkET3A(b *testing.B)     { runBenchmark(b, "ET3A") }
func BenchmarkET3B(b *testing.B)     { runBenchmark(b, "ET3B") }
func BenchmarkET3C(b *testing.B)     { runBenchmark(b, "ET3C") }
func BenchmarkET3D(b *testing.B)     { runBenchmark(b, "ET3D") }
func BenchmarkET3E(b *testing.B)     { runBenchmark(b, "ET3E") }
func BenchmarkET3G(b *testing.B)     { runBenchmark(b, "ET3G") }
func BenchmarkET3H(b *testing.B)     { runBenchmark(b, "ET3H") }
func BenchmarkET3I(b *testing.B)     { runBenchmark(b, "ET3I") }
func BenchmarkET4A(b *testing.B)     { runBenchmark(b, "ET4A") }
func BenchmarkET4B(b *testing.B)     { runBenchmark(b, "ET4B") }
func BenchmarkET4C(b *testing.B)     { runBenchmark(b, "ET4C") }
func BenchmarkET4D(b *testing.B)     { runBenchmark(b, "ET4D") }
func BenchmarkET4E(b *testing.B)     { runBenchmark(b, "ET4E") }
func BenchmarkET4G(b *testing.B)     { runBenchmark(b, "ET4G") }
func BenchmarkET4H(b *testing.B)     { runBenchmark(b, "ET4H") }
func BenchmarkET4I(b *testing.B)     { runBenchmark(b, "ET4I") }
func BenchmarkGOE2A(b *testing.B)    { runBenchmark(b, "GOE2A") }
func BenchmarkGOE2B(b *testing.B)    { runBenchmark(b, "GOE2B") }
func BenchmarkGOE2C(b *testing.B)    { runBenchmark(b, "GOE2C") }
func BenchmarkGOE2E(b *testing.B)    { runBenchmark(b, "GOE2E") }
func BenchmarkGOE3A(b *testing.B)    { runBenchmark(b, "GOE3A") }
func BenchmarkGOE3B(b *testing.B)    { runBenchmark(b, "GOE3B") }
func BenchmarkGOE3C(b *testing.B)    { runBenchmark(b, "GOE3C") }
func BenchmarkGOE3D(b *testing.B)    { runBenchmark(b, "GOE3D") }
func BenchmarkGOE3E(b *testing.B)    { runBenchmark(b, "GOE3E") }
func BenchmarkGOE3G(b *testing.B)    { runBenchmark(b, "GOE3G") }
func BenchmarkGOE3H(b *testing.B)    { runBenchmark(b, "GOE3H") }
func BenchmarkGOE3I(b *testing.B)    { runBenchmark(b, "GOE3I") }
func BenchmarkGOE4A(b *testing.B)    { runBenchmark(b, "GOE4A") }
func BenchmarkGOE4B(b *testing.B)    { runBenchmark(b, "GOE4B") }
func BenchmarkGOE4C(b *testing.B)    { runBenchmark(b, "GOE4C") }
func BenchmarkGOE4D(b *testing.B)    { runBenchmark(b, "GOE4D") }
func BenchmarkGOE4E(b *testing.B)    { runBenchmark(b, "GOE4E") }
func BenchmarkGOE4G(b *testing.B)    { runBenchmark(b, "GOE4G") }
func BenchmarkGOE4H(b *testing.B)    { runBenchmark(b, "GOE4H") }
func BenchmarkGOE4I(b *testing.B)    { runBenchmark(b, "GOE4I") }
func BenchmarkGT2A(b *testing.B)     { runBenchmark(b, "GT2A") }
func BenchmarkGT2B(b *testing.B)     { runBenchmark(b, "GT2B") }
func BenchmarkGT2C(b *testing.B)     { runBenchmark(b, "GT2C") }
func BenchmarkGT2E(b *testing.B)     { runBenchmark(b, "GT2E") }
func BenchmarkGT3A(b *testing.B)     { runBenchmark(b, "GT3A") }
func BenchmarkGT3B(b *testing.B)     { runBenchmark(b, "GT3B") }
func BenchmarkGT3C(b *testing.B)     { runBenchmark(b, "GT3C") }
func BenchmarkGT3D(b *testing.B)     { runBenchmark(b, "GT3D") }
func BenchmarkGT3E(b *testing.B)     { runBenchmark(b, "GT3E") }
func BenchmarkGT3G(b *testing.B)     { runBenchmark(b, "GT3G") }
func BenchmarkGT3H(b *testing.B)     { runBenchmark(b, "GT3H") }
func BenchmarkGT3I(b *testing.B)     { runBenchmark(b, "GT3I") }
func BenchmarkGT4A(b *testing.B)     { runBenchmark(b, "GT4A") }
func BenchmarkGT4B(b *testing.B)     { runBenchmark(b, "GT4B") }
func BenchmarkGT4C(b *testing.B)     { runBenchmark(b, "GT4C") }
func BenchmarkGT4D(b *testing.B)     { runBenchmark(b, "GT4D") }
func BenchmarkGT4E(b *testing.B)     { runBenchmark(b, "GT4E") }
func BenchmarkGT4G(b *testing.B)     { runBenchmark(b, "GT4G") }
func BenchmarkGT4H(b *testing.B)     { runBenchmark(b, "GT4H") }
func BenchmarkGT4I(b *testing.B)     { runBenchmark(b, "GT4I") }
func BenchmarkLOE2A(b *testing.B)    { runBenchmark(b, "LOE2A") }
func BenchmarkLOE2B(b *testing.B)    { runBenchmark(b, "LOE2B") }
func BenchmarkLOE2C(b *testing.B)    { runBenchmark(b, "LOE2C") }
func BenchmarkLOE2E(b *testing.B)    { runBenchmark(b, "LOE2E") }
func BenchmarkLOE3A(b *testing.B)    { runBenchmark(b, "LOE3A") }
func BenchmarkLOE3B(b *testing.B)    { runBenchmark(b, "LOE3B") }
func BenchmarkLOE3C(b *testing.B)    { runBenchmark(b, "LOE3C") }
func BenchmarkLOE3D(b *testing.B)    { runBenchmark(b, "LOE3D") }
func BenchmarkLOE3E(b *testing.B)    { runBenchmark(b, "LOE3E") }
func BenchmarkLOE3G(b *testing.B)    { runBenchmark(b, "LOE3G") }
func BenchmarkLOE3H(b *testing.B)    { runBenchmark(b, "LOE3H") }
func BenchmarkLOE3I(b *testing.B)    { runBenchmark(b, "LOE3I") }
func BenchmarkLOE4A(b *testing.B)    { runBenchmark(b, "LOE4A") }
func BenchmarkLOE4B(b *testing.B)    { runBenchmark(b, "LOE4B") }
func BenchmarkLOE4C(b *testing.B)    { runBenchmark(b, "LOE4C") }
func BenchmarkLOE4D(b *testing.B)    { runBenchmark(b, "LOE4D") }
func BenchmarkLOE4E(b *testing.B)    { runBenchmark(b, "LOE4E") }
func BenchmarkLOE4G(b *testing.B)    { runBenchmark(b, "LOE4G") }
func BenchmarkLOE4H(b *testing.B)    { runBenchmark(b, "LOE4H") }
func BenchmarkLOE4I(b *testing.B)    { runBenchmark(b, "LOE4I") }
func BenchmarkLT2A(b *testing.B)     { runBenchmark(b, "LT2A") }
func BenchmarkLT2B(b *testing.B)     { runBenchmark(b, "LT2B") }
func BenchmarkLT2C(b *testing.B)     { runBenchmark(b, "LT2C") }
func BenchmarkLT2E(b *testing.B)     { runBenchmark(b, "LT2E") }
func BenchmarkLT3A(b *testing.B)     { runBenchmark(b, "LT3A") }
func BenchmarkLT3B(b *testing.B)     { runBenchmark(b, "LT3B") }
func BenchmarkLT3C(b *testing.B)     { runBenchmark(b, "LT3C") }
func BenchmarkLT3D(b *testing.B)     { runBenchmark(b, "LT3D") }
func BenchmarkLT3E(b *testing.B)     { runBenchmark(b, "LT3E") }
func BenchmarkLT3G(b *testing.B)     { runBenchmark(b, "LT3G") }
func BenchmarkLT3H(b *testing.B)     { runBenchmark(b, "LT3H") }
func BenchmarkLT3I(b *testing.B)     { runBenchmark(b, "LT3I") }
func BenchmarkLT4A(b *testing.B)     { runBenchmark(b, "LT4A") }
func BenchmarkLT4B(b *testing.B)     { runBenchmark(b, "LT4B") }
func BenchmarkLT4C(b *testing.B)     { runBenchmark(b, "LT4C") }
func BenchmarkLT4D(b *testing.B)     { runBenchmark(b, "LT4D") }
func BenchmarkLT4E(b *testing.B)     { runBenchmark(b, "LT4E") }
func BenchmarkLT4G(b *testing.B)     { runBenchmark(b, "LT4G") }
func BenchmarkLT4H(b *testing.B)     { runBenchmark(b, "LT4H") }
func BenchmarkLT4I(b *testing.B)     { runBenchmark(b, "LT4I") }
func BenchmarkMax2(b *testing.B)     { runBenchmark(b, "Max2") }
func BenchmarkMax3(b *testing.B)     { runBenchmark(b, "Max3") }
func BenchmarkMax4(b *testing.B)     { runBenchmark(b, "Max4") }
func BenchmarkMin2(b *testing.B)     { runBenchmark(b, "Min2") }
func BenchmarkMin3(b *testing.B)     { runBenchmark(b, "Min3") }
func BenchmarkMin4(b *testing.B)     { runBenchmark(b, "Min4") }
func BenchmarkMul3(b *testing.B)     { runBenchmark(b, "Mul3") }
func BenchmarkMul4(b *testing.B)     { runBenchmark(b, "Mul4") }
func BenchmarkNET2A(b *testing.B)    { runBenchmark(b, "NET2A") }
func BenchmarkNET2B(b *testing.B)    { runBenchmark(b, "NET2B") }
func BenchmarkNET2C(b *testing.B)    { runBenchmark(b, "NET2C") }
func BenchmarkNET2E(b *testing.B)    { runBenchmark(b, "NET2E") }
func BenchmarkNET3A(b *testing.B)    { runBenchmark(b, "NET3A") }
func BenchmarkNET3B(b *testing.B)    { runBenchmark(b, "NET3B") }
func BenchmarkNET3C(b *testing.B)    { runBenchmark(b, "NET3C") }
func BenchmarkNET3D(b *testing.B)    { runBenchmark(b, "NET3D") }
func BenchmarkNET3E(b *testing.B)    { runBenchmark(b, "NET3E") }
func BenchmarkNET3G(b *testing.B)    { runBenchmark(b, "NET3G") }
func BenchmarkNET3H(b *testing.B)    { runBenchmark(b, "NET3H") }
func BenchmarkNET3I(b *testing.B)    { runBenchmark(b, "NET3I") }
func BenchmarkNET4A(b *testing.B)    { runBenchmark(b, "NET4A") }
func BenchmarkNET4B(b *testing.B)    { runBenchmark(b, "NET4B") }
func BenchmarkNET4C(b *testing.B)    { runBenchmark(b, "NET4C") }
func BenchmarkNET4D(b *testing.B)    { runBenchmark(b, "NET4D") }
func BenchmarkNET4E(b *testing.B)    { runBenchmark(b, "NET4E") }
func BenchmarkNET4G(b *testing.B)    { runBenchmark(b, "NET4G") }
func BenchmarkNET4H(b *testing.B)    { runBenchmark(b, "NET4H") }
func BenchmarkNET4I(b *testing.B)    { runBenchmark(b, "NET4I") }
func BenchmarkNeg(b *testing.B)      { runBenchmark(b, "Neg") }
func BenchmarkNop(b *testing.B)      { runBenchmark(b, "Nop") }
func BenchmarkOne(b *testing.B)      { runBenchmark(b, "One") }
func BenchmarkOne2(b *testing.B)     { runBenchmark(b, "One2") }
func BenchmarkSub3(b *testing.B)     { runBenchmark(b, "Sub3") }
func BenchmarkSub4(b *testing.B)     { runBenchmark(b, "Sub4") }
func BenchmarkZero(b *testing.B)     { runBenchmark(b, "Zero") }
func BenchmarkZero2(b *testing.B)    { runBenchmark(b, "Zero2") }
