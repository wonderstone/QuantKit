// Package genomeset provides the basis for a single GEP genomeset.
// A genomeset consists of one or more genomes.
package genomeset

import (
	"log"

	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
)

// GenomeSet contains the genomes that make up the genomeset.
// It also provides the score that results
// from evaluating the GenomeSet against the fitness function.
type GenomeSet struct {
	Genomes  []*genome.Genome `yaml:"genomes,omitempty"`
	LinkFunc string           `yaml:"linkfunc,omitempty"`
	Score    float64          `yaml:"score,omitempty"`
	Index    int
	Iterate  int
}

// NewGenomeSet creates a new GenomeSet from the given genomes.
func New(genomes []*genome.Genome, linkFunc string) *GenomeSet {
	return &GenomeSet{Genomes: genomes, LinkFunc: linkFunc}
}

func (gs *GenomeSet) StringSlice() [][]string {
	var res [][]string
	for _, g := range gs.Genomes {
		res = append(res, g.StringSlice())
	}
	return res
}

// Mutate mutates a genome by performing numMutations random symbol exchanges within the genome.
func (gs *GenomeSet) Mutate(numMutations int) {
	// iter gs.Genomes and mutate each one
	for i := range gs.Genomes {
		gs.Genomes[i].Mutate(numMutations)
	}
}

func (gs *GenomeSet) IsTransposition(gl int) {
	// iter the gs.Genomes and IsTransposition each one
	for i := range gs.Genomes {
		gs.Genomes[i].IsTransposition(gl)
	}
}

func (gs *GenomeSet) RisTransposition(gl int) {
	// iter the gs.Genomes and IsTransposition each one
	for i := range gs.Genomes {
		gs.Genomes[i].RisTransposition(gl)
	}
}

func (gs *GenomeSet) GeneTransposition() {
	// iter the gs.Genomes and GeneTransposition each one
	for i := range gs.Genomes {
		gs.Genomes[i].GeneTransposition()
	}
}

func (gs *GenomeSet) OnePointRecombination(other *GenomeSet) {
	// iter the gs.Genomes and GeneTransposition each one
	for i := range gs.Genomes {
		gs.Genomes[i].OnePointRecombination(other.Genomes[i])
	}
}

func (gs *GenomeSet) TwoPointRecombination(other *GenomeSet) {
	// iter the gs.Genomes and GeneTransposition each one
	for i := range gs.Genomes {
		gs.Genomes[i].TwoPointRecombination(other.Genomes[i])
	}
}

func (gs *GenomeSet) GeneRecombination(other *GenomeSet) {
	// iter the gs.Genomes and GeneTransposition each one
	for i := range gs.Genomes {
		gs.Genomes[i].GeneRecombination(other.Genomes[i])
	}
}

// Dup duplicates the genome into the provided destination genome.
func (gs *GenomeSet) Dup() *GenomeSet {
	if gs == nil {
		log.Printf("denome.Dup error: src and dst must be non-nil")
		return nil

	}
	dst := &GenomeSet{
		Genomes:  make([]*genome.Genome, len(gs.Genomes)),
		LinkFunc: gs.LinkFunc,
		Score:    gs.Score,
	}
	for i := range gs.Genomes {
		dst.Genomes[i] = gs.Genomes[i].Dup()
	}
	return dst
}

// ScoringFunc is the function that is used to evaluate the fitness of the model.
// Typically, a return value of 0 means that the function is nowhere close to being
// a valid solution and a return value of 1000 (or higher) means a perfect solution.
type ScoringFunc func(g *GenomeSet) float64

// EvaluateWithScore scores a genome and sends the result to a channel.
func (g *GenomeSet) EvaluateWithScore(sf ScoringFunc, c chan<- *GenomeSet) {
	if sf == nil {
		log.Fatalf("genome.EvaluateWithScore: ScoringFunc must not be nil")
	}
	g.Score = sf(g)
	c <- g
}
