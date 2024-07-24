// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package genome provides the basis for a single GEP genome.
// A genome consists of one or more genes.
package genome

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/wonderstone/QuantKit/modelgene/gep/gene"
)

// Genome contains the genes that make up the genome.
// It also provides the linking function and the score that results
// from evaluating the genome against the fitness function.
type Genome struct {
	Genes    []*gene.Gene `yaml:"genes,omitempty"`
	LinkFunc string       `yaml:"linkfunc,omitempty"`
	Score    float64      `yaml:"score,omitempty"`
	Index    int          `yaml:"index,omitempty"`
	Iterate  int          `yaml:"iter,omitempty"`

	SymbolMap map[string]int `yaml:"symbolmap,omitempty"` // do not use directly.  Use SymbolCount() instead.
}

// New creates a new genome from the given genes and linking function.
func New(genes []*gene.Gene, linkFunc string) *Genome {
	return &Genome{Genes: genes, LinkFunc: linkFunc}
}

// method to check if the two genomes are equal
func (g *Genome) IfEqual(g2 *Genome) bool {
	if len(g.Genes) != len(g2.Genes) {
		return false
	}
	for i := 0; i < len(g.Genes); i++ {
		if !g.Genes[i].IfEqual(g2.Genes[i]) {
			return false
		}
	}
	// check LinkFunc
	if g.LinkFunc != g2.LinkFunc {
		return false
	}

	return true
}

func merge(dst *map[string]int, src map[string]int) {
	for k, v := range src {
		(*dst)[k] += v
	}
}

// SymbolCount returns the count of the number of times the symbol
// is actually used in the Genome.
// Note that this count is typically different from the number
// of times the symbol appears in the Karva expression.  This can be
// a handy metric to assist in the fitness evaluation of a Genome.
func (g *Genome) SymbolCount(sym string) int {
	if g.SymbolMap == nil {
		g.SymbolMap = make(map[string]int)
		g.SymbolMap[g.LinkFunc] = len(g.Genes) - 1
		for i := 0; i < len(g.Genes); i++ {
			g.Genes[i].SymbolCount(sym) // force evaluation
			m := g.Genes[i].SymbolMap
			merge(&(g.SymbolMap), m)
		}
	}
	return g.SymbolMap[sym]
}

// String returns the Karva representation of the genome.
func (g Genome) String() string {
	result := []string{}
	for _, v := range g.Genes {
		result = append(result, v.String())
	}
	return fmt.Sprintf("%v, score=%v", strings.Join(result, "|"+g.LinkFunc+"|"), g.Score)
}

func (g Genome) StringSlice() []string {
	var result []string
	for _, v := range g.Genes {
		result = append(result, v.String())
	}

	return result
}

// Mutate mutates a genome by performing numMutations random symbol exchanges within the genome.
func (g *Genome) Mutate(numMutations int) {
	for i := 0; i < numMutations; i++ {
		n := rand.Intn(len(g.Genes))
		// fmt.Printf("\nMutating gene #%v, before:\n%v\n", n, g.Genes[n])
		g.Genes[n].Mutate()
		// fmt.Printf("after:\n%v\n", g.Genes[n])
	}
}

// isTransposition transposes the the IS element(start position) to the target site with the length of no more than gl
func (g *Genome) IsTransposition(gl int) {
	// randomly choose the sourceGene
	sourceGene := rand.Intn(len(g.Genes))
	// randomly choose the gene as the target
	targetGene := rand.Intn(len(g.Genes))
	if g.Genes[targetGene].HeadSize != 1 {
		// determine the length of the IS element, the number should be < g.Genes[targetGene].HeadSize
		isLength := rand.Intn(gl) + 1
		if isLength >= g.Genes[targetGene].HeadSize {
			isLength = g.Genes[targetGene].HeadSize - 1
		}
		// the start position of the IS element which should has at least isLength to the end
		startSPosition := rand.Intn(len(g.Genes[sourceGene].Symbols) - isLength)
		// copy IS element out for overlapping reason
		isElement := g.Genes[sourceGene].Symbols[startSPosition : startSPosition+isLength]
		// randomly choose the start position of the IS element which should not be the first position and has at least isLength to the end
		startTPosition := rand.Intn(g.Genes[targetGene].HeadSize - isLength)
		// replace the target with the IS element
		for i := 0; i < isLength; i++ {
			g.Genes[targetGene].Symbols[startTPosition+i+1] = isElement[i]
		}
	}
	// iter the genes to Invalidate!!!! no test
	for i := 0; i < len(g.Genes); i++ {
		g.Genes[i].Invalidate()
	}
}

// risTransposition transposes the the IS function element to the root position with the length of no more than gl
func (g *Genome) RisTransposition(gl int) {
	// randomly choose the sourceGene
	sourceGene := rand.Intn(len(g.Genes))
	// iter the g.Genes[sourceGene].Symbols to find the IS function element and add to a map with Symbol string as key and position int as value
	isMap := make(map[string]int)
	// make a tmp slice for head elements
	head := make([]string, 0)
	// start from the second position of the gene
	for i := 0; i < g.Genes[sourceGene].HeadSize; i++ {
		head = append(head, g.Genes[sourceGene].Symbols[i])
		if i != 0 && !g.Genes[sourceGene].IsTerminal(g.Genes[sourceGene].Symbols[i]) &&
			!g.Genes[sourceGene].IsConstant(g.Genes[sourceGene].Symbols[i]) {
			isMap[g.Genes[sourceGene].Symbols[i]] = i
		}
	}

	// randomly choose the function element from the map
	// map element is randomly arranged ,so each time get the first element is ok
	// iter the map to get a random element
	var tmpElement string
	for k := range isMap {
		tmpElement = k
		break
	}
	isLength := rand.Intn(gl) + 1
	if isLength > g.Genes[sourceGene].HeadSize-isMap[tmpElement] {
		isLength = g.Genes[sourceGene].HeadSize - isMap[tmpElement]
	}
	// copy the RIS element with isLength for overlapping reason
	risElement := g.Genes[sourceGene].Symbols[isMap[tmpElement] : isMap[tmpElement]+isLength]
	// gene.Symmbols rearrange the RIS element to the root position and shift the rest to the right maintaining the order and the length of the gene.Symmbols
	for i := 0; i < g.Genes[sourceGene].HeadSize; i++ {
		if i < len(risElement) {
			g.Genes[sourceGene].Symbols[i] = risElement[i]
		} else {
			g.Genes[sourceGene].Symbols[i] = head[i-len(risElement)]
		}
	}
	// iter the genes to Invalidate!!!! no test
	for i := 0; i < len(g.Genes); i++ {
		g.Genes[i].Invalidate()
	}
}

// GeneTransposition transposes the entire gene funcitons as a whole to the beginning of the chromosome
func (g *Genome) GeneTransposition() {
	// randomly choose the sourceGene but not the first one
	sourceGene := rand.Intn(len(g.Genes)-1) + 1
	// make a temp genome for overlapping
	tmpGenome := g.Dup()

	for i := 0; i < len(g.Genes); i++ {
		if i == 0 {
			g.Genes[i] = tmpGenome.Genes[sourceGene]
		} else if i <= sourceGene {
			g.Genes[i] = tmpGenome.Genes[i-1]
		}
	}
	// iter the genes to Invalidate!!!! no test
	for i := 0; i < len(g.Genes); i++ {
		g.Genes[i].Invalidate()
	}
}

// OnePointRecombination crossover two chromosomes by randomly choosing a point in the chromosome and swapping the genes between the two chromosomes.
func (g *Genome) OnePointRecombination(other *Genome) {
	// randomly choose the sourceGene
	sourceGene := rand.Intn(len(g.Genes))
	// randomly choose the 1 crossoverpoint
	crossoverPoint := rand.Intn(len(g.Genes[sourceGene].Symbols))
	// copy the gene after the crossoverPoint to the tempGeneSlice
	// tempGeneSlice := make([]string, 0)
	for i := crossoverPoint; i < len(g.Genes[sourceGene].Symbols); i++ {
		// tempGeneSlice = append(tempGeneSlice, g.Genes[sourceGene].Symbols[i])
		// exchange g.Genes[sourceGene].Symbols[i] and other.Genes[sourceGene].Symbols[i]
		g.Genes[sourceGene].Symbols[i], other.Genes[sourceGene].Symbols[i] = other.Genes[sourceGene].Symbols[i], g.Genes[sourceGene].Symbols[i]

	}

	// dup the other gene to the tempGenome
	if len(g.Genes) != sourceGene+1 {
		length := len(g.Genes) - sourceGene - 1
		tempGenomeSlice := make([]*gene.Gene, length)
		for i := 0; i < length; i++ {
			tempGenomeSlice[i] = g.Genes[sourceGene+i+1].Dup()
		}

		// transpose the other gene to the g
		for i := sourceGene + 1; i < len(g.Genes); i++ {
			g.Genes[i] = other.Genes[i]
		}
		// transpose the tempGenomeSlice gene to the other
		for i := 0; i < len(tempGenomeSlice); i++ {
			other.Genes[sourceGene+i+1] = tempGenomeSlice[i]
		}
	}
	// iter the genes to Invalidate!!!! no test
	for i := 0; i < len(g.Genes); i++ {
		g.Genes[i].Invalidate()
		other.Genes[i].Invalidate()
	}

}

// the part below  uses different crossover methods（no gene dump process）don't know which one is better
// TwoPointRecombination crossover two chromosomes by randomly choosing two points in the chromosome and swapping the genes between the two chromosomes.
func (g *Genome) TwoPointRecombination(other *Genome) {
	// randomly choose the 2 sourceGenes index. they can be the same
	// assign the smaller index to the first one
	sourceGene1 := rand.Intn(len(g.Genes))
	sourceGene2 := rand.Intn(len(g.Genes))
	if sourceGene1 > sourceGene2 {
		sourceGene1, sourceGene2 = sourceGene2, sourceGene1
	}
	// randomly choose the 2 crossoverpoints
	// if 2 sourceGenes are the same, the crossoverpoints must be the different
	var crossoverPoint1, crossoverPoint2 int
	if sourceGene1 == sourceGene2 {
		crossoverPoint1 = rand.Intn(len(g.Genes[sourceGene1].Symbols))
		for {
			crossoverPoint2 = rand.Intn(len(g.Genes[sourceGene2].Symbols))
			if crossoverPoint2 != crossoverPoint1 {
				break
			}
		}
	} else {
		crossoverPoint1 = rand.Intn(len(g.Genes[sourceGene1].Symbols))
		crossoverPoint2 = rand.Intn(len(g.Genes[sourceGene2].Symbols))
	}
	// make crossoverPoint1 smaller than crossoverPoint2
	if crossoverPoint1 > crossoverPoint2 {
		crossoverPoint1, crossoverPoint2 = crossoverPoint2, crossoverPoint1
	}

	if sourceGene1 == sourceGene2 {
		for i := sourceGene1; i <= sourceGene2; i++ {
			for j := crossoverPoint1; j < crossoverPoint2; j++ {
				// exchange the g.Genes[i].Symbols[j] and other.Genes[i].Symbols[j] at i,j
				g.Genes[i].Symbols[j], other.Genes[i].Symbols[j] = other.Genes[i].Symbols[j], g.Genes[i].Symbols[j]
			}
		}
	} else {
		for i := sourceGene1; i <= sourceGene2; i++ {
			if i == sourceGene1 {
				for j := crossoverPoint1; j < len(g.Genes[i].Symbols); j++ {
					// exchange the g.Genes[i].Symbols[j] and other.Genes[i].Symbols[j] at i,j
					g.Genes[i].Symbols[j], other.Genes[i].Symbols[j] = other.Genes[i].Symbols[j], g.Genes[i].Symbols[j]
				}
			} else if i == sourceGene2 {
				for j := 0; j < crossoverPoint2; j++ {
					// exchange the g.Genes[i].Symbols[j] and other.Genes[i].Symbols[j] at i,j
					g.Genes[i].Symbols[j], other.Genes[i].Symbols[j] = other.Genes[i].Symbols[j], g.Genes[i].Symbols[j]
				}
			} else {
				for j := 0; j < len(g.Genes[i].Symbols); j++ {
					// exchange the g.Genes[i].Symbols[j] and other.Genes[i].Symbols[j] at i,j
					g.Genes[i].Symbols[j], other.Genes[i].Symbols[j] = other.Genes[i].Symbols[j], g.Genes[i].Symbols[j]
				}
			}
		}
	}
	// iter the genes to Invalidate!!!! no test
	for i := 0; i < len(g.Genes); i++ {
		g.Genes[i].Invalidate()
		other.Genes[i].Invalidate()
	}
}

// GeneRecombination crossover two chromosomes by randomly choosing two genes in the chromosome and swapping the genes between the two chromosomes.
func (g *Genome) GeneRecombination(other *Genome) {
	// randomly choose the genePosition for g and other
	genePosition := rand.Intn(len(g.Genes))
	// create a temp gene to store the dup
	tempGene := g.Genes[genePosition].Dup()
	g.Genes[genePosition] = other.Genes[genePosition].Dup()
	other.Genes[genePosition] = tempGene.Dup()
	// iter the genes to Invalidate!!!! no test
	for i := 0; i < len(g.Genes); i++ {
		g.Genes[i].Invalidate()
		other.Genes[i].Invalidate()
	}
}

// Dup duplicates the genome into the provided destination genome.
func (g *Genome) Dup() *Genome {
	if g == nil {
		log.Printf("denome.Dup error: src and dst must be non-nil")
		return nil

	}
	dst := &Genome{
		Genes:    make([]*gene.Gene, len(g.Genes)),
		LinkFunc: g.LinkFunc,
		Score:    g.Score,
		Index:    g.Index,
		Iterate:  g.Iterate,
	}
	for i := range g.Genes {
		dst.Genes[i] = g.Genes[i].Dup()
	}
	return dst
}

// ScoringFunc is the function that is used to evaluate the fitness of the model.
// Typically, a return value of 0 means that the function is nowhere close to being
// a valid solution and a return value of 1000 (or higher) means a perfect solution.
type ScoringFunc func(g *Genome) float64

type IteratorOnceFunc func()

// EvaluateWithScore scores a genome and sends the result to a channel.
func (g *Genome) EvaluateWithScore(sf ScoringFunc, c chan<- *Genome) {
	if sf == nil {
		log.Fatalf("genome.EvaluateWithScore: ScoringFunc must not be nil")
	}
	g.Score = sf(g)
	c <- g
}

// Evaluate runs the model with the observation and populates the provided action
// based on the link function.
// func (g *Genome) Evaluate(stepsSinceReset int, obs gym.Obs, action interface{}) error {
// 	result := make([]int, len(g.Genes))
// 	var in int
// 	if err := obs.Unmarshal(&in); err != nil {
// 		return fmt.Errorf("Unmarshal: %v", err)
// 	}
// 	g.EvalIntTuple([]int{in, stepsSinceReset}, result)
// 	switch v := action.(type) {
// 	case *[]int:
// 		for _, val := range result {
// 			*v = append(*v, val)
// 		}
// 	default:
// 		return fmt.Errorf("Action type %v not yet supported", v)
// 	}
// 	return nil
// }
