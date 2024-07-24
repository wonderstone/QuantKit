// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package gene

import (
	"log"
	"strconv"

	mn "github.com/wonderstone/QuantKit/modelgene/gep/functions/math_nodes"
)

func (g *Gene) generateMathFunc() {
	argOrder := g.getArgOrder()
	g.SymbolMap = make(map[string]int)
	g.mf = g.buildMathTree(0, argOrder)
}

func (g *Gene) GenerateMathFuncVars() *[]string {
	argOrder := g.getArgOrder()
	g.SymbolMap = make(map[string]int)
	temp, vars := g.BuildMathTreeVars(0, argOrder, &[]string{})
	g.mf = temp
	return vars
}

// EvalMath evaluates the gene as a floating-point expression and returns the result.
// in represents the float64 inputs available to the gene.
func (g *Gene) EvalMath(in []float64) float64 {
	if g.mf == nil {
		g.generateMathFunc()
	}
	return g.mf(in)
}

func (g *Gene) buildMathTree(symbolIndex int, argOrder [][]int) func([]float64) float64 {
	// count := make(map[string]int)
	// log.Infof("buildMathTree(%v, %#v, ...)", symbolIndex, argOrder)
	if symbolIndex > len(g.Symbols) {
		log.Printf("bad symbolIndex %v for symbols: %v", symbolIndex, g.Symbols)
		return func(a []float64) float64 { return 0.0 }
	}
	sym := g.Symbols[symbolIndex]
	g.SymbolMap[sym]++
	if s, ok := mn.Math[sym]; ok {
		args := argOrder[symbolIndex]
		var funcs []func([]float64) float64
		for _, arg := range args {
			f := g.buildMathTree(arg, argOrder)
			funcs = append(funcs, f)
		}
		return func(in []float64) float64 {
			var values []float64
			for _, f := range funcs {
				values = append(values, f(in))
			}
			return s.Float64Function(values)
		}
	} else { // No named symbol found - look for d0, d1, ...
		if sym[0:1] == "d" {
			if index, err := strconv.Atoi(sym[1:]); err != nil {
				log.Printf("unable to parse variable index: sym=%q", sym)
			} else {
				return func(in []float64) float64 {
					if index >= len(in) {
						log.Printf("error evaluating gene %q: index %v >= d length (%v)", sym, index, len(in))
						return 0.0
					}
					return in[index]
				}
			}
		} else if sym[0:1] == "c" {
			if index, err := strconv.Atoi(sym[1:]); err != nil {
				log.Printf("unable to parse constant index: sym=%v", sym)
			} else {
				return func(in []float64) float64 {
					if index >= len(g.Constants) {
						log.Printf("error evaluating gene %q: index %v >= c length (%v)", sym, index, len(g.Constants))
						return 0.0
					}
					return g.Constants[index]
				}
			}
		}
	}
	log.Printf("unable to return function: unknown gene symbol %q", sym)
	return func(in []float64) float64 { return 0.0 }
}

func (g *Gene) BuildMathTreeVars(symbolIndex int, argOrder [][]int, vars *[]string) (func([]float64) float64, *[]string) {
	// count := make(map[string]int)
	// log.Infof("buildMathTree(%v, %#v, ...)", symbolIndex, argOrder)
	if symbolIndex > len(g.Symbols) {
		log.Printf("bad symbolIndex %v for symbols: %v", symbolIndex, g.Symbols)
		return func(a []float64) float64 { return 0.0 }, vars
	}
	sym := g.Symbols[symbolIndex]
	g.SymbolMap[sym]++
	if s, ok := mn.Math[sym]; ok {
		args := argOrder[symbolIndex]
		var funcs []func([]float64) float64
		for _, arg := range args {
			f, tempvars := g.BuildMathTreeVars(arg, argOrder, vars)
			funcs = append(funcs, f)
			vars = tempvars
		}
		return func(in []float64) float64 {
			var values []float64
			for _, f := range funcs {
				values = append(values, f(in))
			}
			return s.Float64Function(values)
		}, vars
	} else { // No named symbol found - look for d0, d1, ...
		if sym[0:1] == "d" {
			if index, err := strconv.Atoi(sym[1:]); err != nil {
				log.Printf("unable to parse variable index: sym=%q", sym)
			} else {
				tmpf := func(in []float64) float64 {
					if index >= len(in) {
						log.Printf("error evaluating gene %q: index %v >= d length (%v)", sym, index, len(in))
						return 0.0
					}

					return in[index]
				}
				*vars = append(*vars, sym)
				return tmpf, vars
			}
		} else if sym[0:1] == "c" {
			if index, err := strconv.Atoi(sym[1:]); err != nil {
				log.Printf("unable to parse constant index: sym=%v", sym)
			} else {
				tmpf := func(in []float64) float64 {
					if index >= len(g.Constants) {
						log.Printf("error evaluating gene %q: index %v >= c length (%v)", sym, index, len(g.Constants))
						return 0.0
					}

					return g.Constants[index]
				}
				*vars = append(*vars, sym)
				return tmpf, vars
			}
		}
	}
	log.Printf("unable to return function: unknown gene symbol %q", sym)
	return func(in []float64) float64 { return 0.0 }, vars
}

// the reason why we did not use the type void struct{}
func RemoveDuplicates(slice []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range slice {
		if !encountered[slice[v]] {

			// Record this element as an encountered element.
			encountered[slice[v]] = true
			// Append to result slice.
			result = append(result, slice[v])
		}
	}
	// Return the new slice.
	return result
}
