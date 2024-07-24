// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package model provides the complete representation of the model for a given GEP problem.
package model

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/wonderstone/QuantKit/modelgene/gep/functions"
	bn "github.com/wonderstone/QuantKit/modelgene/gep/functions/bool_nodes"
	in "github.com/wonderstone/QuantKit/modelgene/gep/functions/int_nodes"
	mn "github.com/wonderstone/QuantKit/modelgene/gep/functions/math_nodes"
	"github.com/wonderstone/QuantKit/modelgene/gep/gene"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
)

// Generation represents one complete generation of the model.
type Generation struct {
	Genomes     []*genome.Genome
	Funcs       []gene.FuncWeight
	ScoringFunc genome.ScoringFunc
}

// New creates a new random generation of the model.
// fs is a slice of function weights.
// funcType is the underlying function type (no generics).
// numGenomes is the number of genomes to use to populate this generation of the model.
// headSize is the number of head symbols to use in a genome.
// numGenesPerGenome is the number of genes to use per genome.
// numTerminals is the number of terminals (inputs) to use within each gene.
// numConstants is the number of constants (inputs) to use within each gene.
// linkFunc is the linking function used to combine the genes within a genome.
// sf is the scoring (or fitness) function.
func New(
	fs []gene.FuncWeight, funcType functions.FuncType,
	numGenomes, headSize, numGenesPerGenome, numTerminals, numConstants int, linkFunc string, sf genome.ScoringFunc,
) *Generation {
	r := &Generation{
		Genomes:     make([]*genome.Genome, numGenomes),
		Funcs:       fs,
		ScoringFunc: sf,
	}
	n := maxArity(fs, funcType)
	tailSize := headSize*(n-1) + 1
	for i := range r.Genomes {
		genes := make([]*gene.Gene, numGenesPerGenome)
		for j := range genes {
			genes[j] = gene.RandomNew(headSize, tailSize, numTerminals, numConstants, fs, funcType)
		}
		r.Genomes[i] = genome.New(genes, linkFunc)
	}
	return r
}

// Evolve runs the GEP algorithm for the given number of iterations, or until a score of expectFitness (or more) is reached.
func (g *Generation) Evolve(
	iterations int, expectFitness float64, pm float64, pis float64, glis int, pris float64, glris int, pgene float64,
	p1p float64, p2p float64, pr float64,
) *genome.Genome {
	// Algorithm flow diagram, figure 3.1, book page 56
	for i := 0; i < iterations; i++ {

		bestGenome := g.getBest() // Preserve the best genome
		// fmt.Println("Iteration #", i, bestGenome.Score)
		if bestGenome.Score >= expectFitness {
			fmt.Printf("Stopping after generation #%v\n", i)
			return bestGenome
		}
		// fmt.Printf("Best genome (score %v): %v\n", bestGenome.Score, *bestGenome)
		saveCopy := bestGenome.Dup()
		g.replication() // Section 3.3.1, book page 75
		if pm <= 0 || pm >= 1 {
			panic("pm must be between 0 and 1")
		} else {
			g.mutation(pm) // Section 3.3.2, book page 77
		}
		if pis < 0 || pis >= 1 {
			panic("pis must be between 0 and 1")
		} else {
			g.isTransposition(pis, glis)
		}
		if pris < 0 || pris >= 1 {
			panic("pris must be between 0 and 1")
		} else {
			g.risTransposition(pris, glris)
		}
		if pgene < 0 || pgene >= 1 {
			panic("pgene must be between 0 and 1")
		} else {
			g.geneTransposition(pgene)
		}
		if p1p < 0 || p1p >= 1 {
			panic("p1p must be between 0 and 1")
		} else {
			g.onePointRecombination(p1p)
		}
		if p2p < 0 || p2p >= 1 {
			panic("p2p must be between 0 and 1")
		} else {
			g.twoPointRecombination(p2p)
		}
		if pr < 0 || pr >= 1 {
			panic("pr must be between 0 and 1")
		} else {
			g.geneRecombination(pr)
		}
		// Now that replication is done, restore the best genome (aka "elitism")
		g.Genomes[0] = saveCopy
		// fmt.Printf("the round right now is #%v, best score is %v\n", i, bestGenome.Score)
	}
	// fmt.Printf("Stopping after generation #%v\n", iterations)
	return g.getBest()
}

func (g *Generation) replication() {
	if len(g.Genomes) == 0 {
		return
	}

	maxWeight := g.Genomes[0].Score
	minWeight := g.Genomes[0].Score
	for _, v := range g.Genomes {
		maxWeight = math.Max(maxWeight, v.Score)
		minWeight = math.Min(minWeight, v.Score)
	}

	maxWeight = maxWeight - minWeight + 1.0
	result := make([]*genome.Genome, 0, len(g.Genomes))
	index := rand.Intn(len(g.Genomes))
	beta := 0.0
	for i := 0; i < len(g.Genomes); i++ {
		beta += rand.Float64() * 2.0 * maxWeight
		for beta > g.Genomes[index].Score-minWeight+1.0 {
			beta -= g.Genomes[index].Score - minWeight + 1.0
			index = (index + 1) % len(g.Genomes)
		}
		result = append(result, g.Genomes[index].Dup())
	}
	g.Genomes = result
}

func (g *Generation) mutation(pm float64) {
	// Determine the total number of genomes to mutate
	// numGenomes := 1 + rand.Intn(len(g.Genomes)-1)
	numGenomes := int(math.Ceil(float64(len(g.Genomes)) * pm))
	for i := 0; i < numGenomes; i++ {
		// Pick a random genome
		genomeNum := rand.Intn(len(g.Genomes))
		gen := g.Genomes[genomeNum]
		// Determine the total number of mutations to perform within the genome
		// Do not change this part, respect the init coder
		numMutations := 1 + rand.Intn(2)
		// fmt.Printf("\nMutating genome #%v %v times, before:\n%v\n", genomeNum, numMutations, genome)
		gen.Mutate(numMutations)
		// fmt.Printf("after:\n%v\n", genome)
	}
}

func (g *Generation) isTransposition(pis float64, gl int) {
	// Determine the total number of genomes to isTransposition with pis upround
	numGenomes := int(math.Ceil(float64(len(g.Genomes)) * pis))
	for i := 0; i < numGenomes; i++ {
		// Pick a random genome
		genomeNum := rand.Intn(len(g.Genomes))
		gen := g.Genomes[genomeNum]
		// Perform the isTransposition within the genome
		gen.IsTransposition(gl)

	}
}

func (g *Generation) risTransposition(pris float64, gl int) {
	// Determine the total number of genomes to risTransposition with pris upround
	numGenomes := int(math.Ceil(float64(len(g.Genomes)) * pris))
	for i := 0; i < numGenomes; i++ {
		// Pick a random genome
		genomeNum := rand.Intn(len(g.Genomes))
		gen := g.Genomes[genomeNum]
		// Perform the risTransposition within the genome
		gen.RisTransposition(gl)
	}
}

func (g *Generation) geneTransposition(pgene float64) {
	// Determine the total number of genomes to geneTransposition with pgene upround
	numGenomes := int(math.Ceil(float64(len(g.Genomes)) * pgene))
	for i := 0; i < numGenomes; i++ {
		// Pick a random genome
		genomeNum := rand.Intn(len(g.Genomes))
		gen := g.Genomes[genomeNum]
		// Perform the geneTransposition within the genome
		gen.GeneTransposition()
	}
}

func (g *Generation) onePointRecombination(p1p float64) {
	// Determine the total number of genomes to onePointRecombination with p1p upround
	numGenomes := int(math.Ceil(float64(len(g.Genomes)) * p1p))
	for i := 0; i < numGenomes; i++ {
		// Pick two different random genomes
		genomeNum1 := rand.Intn(len(g.Genomes))
		var genomeNum2 int
		for {
			genomeNum2 = rand.Intn(len(g.Genomes))
			if genomeNum1 != genomeNum2 {
				break
			}
		}
		gen1 := g.Genomes[genomeNum1]
		gen2 := g.Genomes[genomeNum2]
		// Perform the onePointRecombination within the genome
		gen1.OnePointRecombination(gen2)
	}
}

func (g *Generation) twoPointRecombination(p2p float64) {
	// Determine the total number of genomes to twoPointRecombination with p2p upround
	numGenomes := int(math.Ceil(float64(len(g.Genomes)) * p2p))
	for i := 0; i < numGenomes; i++ {
		// Pick two different random genomes
		genomeNum1 := rand.Intn(len(g.Genomes))
		var genomeNum2 int
		for {
			genomeNum2 = rand.Intn(len(g.Genomes))
			if genomeNum1 != genomeNum2 {
				break
			}
		}
		gen1 := g.Genomes[genomeNum1]
		gen2 := g.Genomes[genomeNum2]
		// Perform the twoPointRecombination within the genome
		gen1.TwoPointRecombination(gen2)
	}
}

func (g *Generation) geneRecombination(pr float64) {
	// Determine the total number of genomes to geneRecombination with pr upround
	numGenomes := int(math.Ceil(float64(len(g.Genomes)) * pr))
	for i := 0; i < numGenomes; i++ {
		// Pick two different random genomes
		genomeNum1 := rand.Intn(len(g.Genomes))
		var genomeNum2 int
		for {
			genomeNum2 = rand.Intn(len(g.Genomes))
			if genomeNum1 != genomeNum2 {
				break
			}
		}
		gen1 := g.Genomes[genomeNum1]
		gen2 := g.Genomes[genomeNum2]
		// Perform the geneRecombination within the genome
		gen1.GeneRecombination(gen2)
	}
}

// getBest evaluates all genomes and returns a pointer to the best one.
func (g *Generation) getBest() *genome.Genome {
	bestScore := 0.0
	bestGenome := g.Genomes[0]
	c := make(chan *genome.Genome)
	for i := 0; i < len(g.Genomes); i++ { // Evaluate genomes concurrently
		go g.Genomes[i].EvaluateWithScore(g.ScoringFunc, c)
	}
	for i := 0; i < len(g.Genomes); i++ { // Collect and return the highest scoring Genome
		gn := <-c
		if gn.Score > bestScore {
			bestGenome = gn
			bestScore = gn.Score
		}
	}
	return bestGenome
}

// maxArity determines the maximum number of input terminals for the given set of symbols.
func maxArity(fs []gene.FuncWeight, funcType functions.FuncType) int {
	var lookup functions.FuncMap
	switch funcType {
	case functions.Bool:
		lookup = bn.BoolAllGates
	case functions.Int:
		lookup = in.Int
	case functions.Float64:
		lookup = mn.Math
	default:
		log.Fatalf("unknown funcType: %v", funcType)
	}

	r := 0
	for _, f := range fs {
		if fn, ok := lookup[f.Symbol]; ok {
			if fn.Terminals() > r {
				r = fn.Terminals()
			}
		} else {
			log.Printf("unable to find symbol %v in function map", f.Symbol)
		}
	}
	return r
}

// GenerationGS represents one complete generation of the modelGS.
type GenerationGS struct {
	GenomeSets  []*genomeset.GenomeSet
	Funcs       []gene.FuncWeight
	ScoringFunc genomeset.ScoringFunc
}

// New creates a new random generation of the model.
// fs is a slice of function weights.
// funcType is the underlying function type (no generics).
// num
// numGenomeSets is the number of genomeset to use to populate this generation of the modelgs.
// headSize is the number of head symbols to use in a genome.
// numGenomesPerGenomeSet is the number of genomes to use per genomeset.
// numGenesPerGenome is the number of genes to use per genome.
// numTerminals is the number of terminals (inputs) to use within each gene.
// numConstants is the number of constants (inputs) to use within each gene.
// linkFunc is the linking function used to combine the genes within a genome.
// sf is the scoring (or fitness) function.
func NewGS(
	fs []gene.FuncWeight, funcType functions.FuncType,
	numGenomeSets, headSize, numGenomesPerGenomeSet, numGenesPerGenome, numTerminals, numConstants int, linkFunc string,
	sf genomeset.ScoringFunc,
) *GenerationGS {
	r := &GenerationGS{
		GenomeSets:  make([]*genomeset.GenomeSet, numGenomeSets),
		Funcs:       fs,
		ScoringFunc: sf,
	}
	n := maxArity(fs, funcType)
	tailSize := headSize*(n-1) + 1
	for i := range r.GenomeSets {
		r.GenomeSets[i] = genomeset.New(make([]*genome.Genome, numGenomesPerGenomeSet), linkFunc)
		for j := range r.GenomeSets[i].Genomes {
			genes := make([]*gene.Gene, numGenesPerGenome)
			for k := range genes {
				genes[k] = gene.RandomNew(headSize, tailSize, numTerminals, numConstants, fs, funcType)
			}
			r.GenomeSets[i].Genomes[j] = genome.New(genes, linkFunc)
		}
	}
	return r
}

// EvolveGS runs the GEP algorithm for the given number of iterations, or until a score of expectFitness (or more) is reached.
func (gs *GenerationGS) EvolveGS(
	iterations int, expectFitness float64, pm float64, pis float64, glis int, pris float64, glris int, pgene float64,
	p1p float64, p2p float64, pr float64,
) *genomeset.GenomeSet {
	// Algorithm flow diagram, figure 3.1, book page 56
	for i := 0; i < iterations; i++ {
		// fmt.Printf("Iteration #%v...\n", i)
		bestGenome := gs.getBest() // Preserve the best genome
		if bestGenome.Score >= expectFitness {
			fmt.Printf("Stopping after generation #%v\n", i)
			return bestGenome
		}
		// fmt.Printf("Best genome (score %v): %v\n", bestGenome.Score, *bestGenome)
		saveCopy := bestGenome.Dup()
		gs.replication() // Section 3.3.1, book page 75
		if pm <= 0 || pm >= 1 {
			panic("pm must be between 0 and 1")
		} else {
			gs.mutation(pm) // Section 3.3.2, book page 77
		}
		if pis <= 0 || pis >= 1 {
			panic("pis must be between 0 and 1")
		} else {
			gs.isTransposition(pis, glis)
		}
		if pris <= 0 || pris >= 1 {
			panic("pris must be between 0 and 1")
		} else {
			gs.risTransposition(pris, glris)
		}
		if pgene <= 0 || pgene >= 1 {
			panic("pgene must be between 0 and 1")
		} else {
			gs.geneTransposition(pgene)
		}
		if p1p <= 0 || p1p >= 1 {
			panic("p1p must be between 0 and 1")
		} else {
			gs.onePointRecombination(p1p)
		}
		if p2p <= 0 || p2p >= 1 {
			panic("p2p must be between 0 and 1")
		} else {
			gs.twoPointRecombination(p2p)
		}
		if pr <= 0 || pr >= 1 {
			panic("pr must be between 0 and 1")
		} else {
			gs.geneRecombination(pr)
		}
		// Now that replication is done, restore the best genome (aka "elitism")
		gs.GenomeSets[0] = saveCopy
		// fmt.Printf("the round right now is #%v, best score is %v\n", i, bestGenome.Score)
	}
	// fmt.Printf("Stopping after generation #%v\n", iterations)
	return gs.getBest()
}

func (gs *GenerationGS) replication() {
	if len(gs.GenomeSets) == 0 {
		return
	}

	maxWeight := gs.GenomeSets[0].Score
	minWeight := gs.GenomeSets[0].Score
	for _, v := range gs.GenomeSets {
		maxWeight = math.Max(maxWeight, v.Score)
		minWeight = math.Min(minWeight, v.Score)
	}

	maxWeight = maxWeight - minWeight + 1.0
	result := make([]*genomeset.GenomeSet, 0, len(gs.GenomeSets))
	index := rand.Intn(len(gs.GenomeSets))
	beta := 0.0
	for i := 0; i < len(gs.GenomeSets); i++ {
		beta += rand.Float64() * 2.0 * maxWeight
		for beta > gs.GenomeSets[index].Score-minWeight+1.0 {
			beta -= gs.GenomeSets[index].Score - minWeight + 1.0
			index = (index + 1) % len(gs.GenomeSets)
		}
		result = append(result, gs.GenomeSets[index].Dup())
	}
	gs.GenomeSets = result
}

func (gs *GenerationGS) mutation(pm float64) {
	// Determine the total number of genomeSets to mutate
	numGenomesets := int(math.Ceil(float64(len(gs.GenomeSets)) * pm))

	for i := 0; i < numGenomesets; i++ {
		// Pick a random genomeSet
		genomeSetNum := rand.Intn(len(gs.GenomeSets))
		genSet := gs.GenomeSets[genomeSetNum]
		// Determine the total number of mutations to perform within the genome
		numMutations := 1 + rand.Intn(2)
		// fmt.Printf("\nMutating genome #%v %v times, before:\n%v\n", genomeNum, numMutations, genome)
		genSet.Mutate(numMutations)
		// fmt.Printf("after:\n%v\n", genome)
	}
}

func (gs *GenerationGS) isTransposition(pis float64, gl int) {
	// Determine the total number of genomeSets to isTransposition with pis upround
	numGenomesets := int(math.Ceil(float64(len(gs.GenomeSets)) * pis))
	for i := 0; i < numGenomesets; i++ {
		// Pick a random genomeSet
		genomeSetNum := rand.Intn(len(gs.GenomeSets))
		genSet := gs.GenomeSets[genomeSetNum]
		// Perform the isTransposition within the genomeset
		genSet.IsTransposition(gl)
	}
}

func (gs *GenerationGS) risTransposition(pris float64, gl int) {
	// Determine the total number of genomeSets to risTransposition with pris upround
	numGenomesets := int(math.Ceil(float64(len(gs.GenomeSets)) * pris))
	for i := 0; i < numGenomesets; i++ {
		// Pick a random genomeSet
		genomeSetNum := rand.Intn(len(gs.GenomeSets))
		genSet := gs.GenomeSets[genomeSetNum]
		// Perform the risTransposition within the genomeset
		genSet.RisTransposition(gl)
	}
}

func (gs *GenerationGS) geneTransposition(pgene float64) {
	// Determine the total number of genomeSets to geneTransposition with pgene upround
	numGenomesets := int(math.Ceil(float64(len(gs.GenomeSets)) * pgene))
	for i := 0; i < numGenomesets; i++ {
		// Pick a random genomeSet
		genomeSetNum := rand.Intn(len(gs.GenomeSets))
		genSet := gs.GenomeSets[genomeSetNum]
		// Perform the geneTransposition within the genomeset
		genSet.GeneTransposition()
	}
}

func (gs *GenerationGS) onePointRecombination(p1p float64) {
	// Determine the total number of genomeSets to onePointRecombination with p1p upround
	numGenomesets := int(math.Ceil(float64(len(gs.GenomeSets)) * p1p))
	for i := 0; i < numGenomesets; i++ {
		// pick two different random genomeSets
		genomeSetNum1 := rand.Intn(len(gs.GenomeSets))
		var genomeSetNum2 int
		for {
			genomeSetNum2 = rand.Intn(len(gs.GenomeSets))
			if genomeSetNum1 != genomeSetNum2 {
				break
			}
		}
		genSet1 := gs.GenomeSets[genomeSetNum1]
		genSet2 := gs.GenomeSets[genomeSetNum2]
		// Perform the onePointRecombination within the genomeset
		genSet1.OnePointRecombination(genSet2)
	}
}

func (gs *GenerationGS) twoPointRecombination(p2p float64) {
	// Determine the total number of genomeSets to twoPointRecombination with p2p upround
	numGenomesets := int(math.Ceil(float64(len(gs.GenomeSets)) * p2p))
	for i := 0; i < numGenomesets; i++ {
		// pick two different random genomeSets
		genomeSetNum1 := rand.Intn(len(gs.GenomeSets))
		var genomeSetNum2 int
		for {
			genomeSetNum2 = rand.Intn(len(gs.GenomeSets))
			if genomeSetNum1 != genomeSetNum2 {
				break
			}
		}
		genSet1 := gs.GenomeSets[genomeSetNum1]
		genSet2 := gs.GenomeSets[genomeSetNum2]
		// Perform the twoPointRecombination within the genomeset
		genSet1.TwoPointRecombination(genSet2)
	}
}

func (gs *GenerationGS) geneRecombination(pr float64) {
	// Determine the total number of genomeSets to geneRecombination with pr upround
	numGenomesets := int(math.Ceil(float64(len(gs.GenomeSets)) * pr))
	for i := 0; i < numGenomesets; i++ {
		// pick two different random genomeSets
		genomeSetNum1 := rand.Intn(len(gs.GenomeSets))
		var genomeSetNum2 int
		for {
			genomeSetNum2 = rand.Intn(len(gs.GenomeSets))
			if genomeSetNum1 != genomeSetNum2 {
				break
			}
		}
		genSet1 := gs.GenomeSets[genomeSetNum1]
		genSet2 := gs.GenomeSets[genomeSetNum2]
		// Perform the geneRecombination within the genomeset
		genSet1.GeneRecombination(genSet2)
	}
}

// getBest evaluates all genomes and returns a pointer to the best one.
func (gs *GenerationGS) getBest() *genomeset.GenomeSet {
	bestScore := 0.0
	bestGenomeSet := gs.GenomeSets[0]
	c := make(chan *genomeset.GenomeSet)
	for i := 0; i < len(gs.GenomeSets); i++ { // Evaluate genomes concurrently
		go gs.GenomeSets[i].EvaluateWithScore(gs.ScoringFunc, c)
	}
	for i := 0; i < len(gs.GenomeSets); i++ { // Collect and return the highest scoring Genome
		gn := <-c
		if gn.Score > bestScore {
			bestGenomeSet = gn
			bestScore = gn.Score
		}
	}
	return bestGenomeSet
}
