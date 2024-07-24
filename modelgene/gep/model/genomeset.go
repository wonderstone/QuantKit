package model

import (
	"math"
	"math/rand"

	"github.com/wonderstone/QuantKit/modelgene/gep/gene"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
	"github.com/wonderstone/QuantKit/config"
)

type GenomeSetHandler struct {
	Op         *Op
	GenomeSets []*genomeset.GenomeSet
	GenomeSet  *genomeset.GenomeSet
	Funcs      []gene.FuncWeight
}

func (g *GenomeSetHandler) Init(option ...WithOption) error {
	g.Op = NewOp(option...)

	if g.Op.NumTerminal <= 0 {
		config.ErrorF("输入的参数变量 必须大于0，请利用 SetGEPInputParams 设置")
		return nil
	}

	var functions []gene.FuncWeight
	for _, f := range g.Op.Conf.Function {
		functions = append(
			functions, gene.FuncWeight{
				Symbol: f.Symbol,
				Weight: f.Weight,
			},
		)
	}

	if g.Op.GenomeSetF != nil {
		g.Funcs = functions
		g.GenomeSet = genomeset.New(
			make([]*genome.Genome, g.Op.Conf.NumGenomesPerGenomeSet), g.Op.Conf.LinkFunc,
		)

		for j := range g.GenomeSet.Genomes {
			genes := make([]*gene.Gene, g.Op.Conf.NumGenesPerGenome)
			for k := range genes {
				genes[k] = gene.New(g.Op.Record.Gep.KES[j][k], g.Op.FuncType)
			}
			g.GenomeSet.Genomes[j] = genome.New(genes, g.Op.Conf.LinkFunc)
		}

	} else {
		g.GenomeSets = make([]*genomeset.GenomeSet, g.Op.Conf.NumGenomes)
		g.Funcs = functions
		n := maxArity(functions, g.Op.FuncType)
		tailSize := g.Op.Conf.HeadSize*(n-1) + 1
		for i := range g.GenomeSets {
			g.GenomeSets[i] = genomeset.New(
				make([]*genome.Genome, g.Op.Conf.NumGenomesPerGenomeSet), g.Op.Conf.LinkFunc,
			)
			for j := range g.GenomeSets[i].Genomes {
				genes := make([]*gene.Gene, g.Op.Conf.NumGenesPerGenome)
				for k := range genes {
					genes[k] = gene.RandomNew(
						g.Op.Conf.HeadSize, tailSize, g.Op.NumTerminal, g.Op.Conf.NumConstants, functions,
						g.Op.FuncType,
					)
				}
				g.GenomeSets[i].Genomes[j] = genome.New(genes, g.Op.Conf.LinkFunc)
			}
		}
	}

	return nil
}

func (g *GenomeSetHandler) makeRecord(gene *genomeset.GenomeSet) *Record {
	// 将karva表达式写入文件
	return &Record{
		Gep: GepRecord{
			Mode:  "GenomeSet",
			Score: gene.Score,
			KES:   gene.StringSlice(),
		},
	}
}
func (gs *GenomeSetHandler) replication() {
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

func (gs *GenomeSetHandler) mutation(pm float64) {
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

func (gs *GenomeSetHandler) isTransposition(pis float64, gl int) {
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

func (gs *GenomeSetHandler) risTransposition(pris float64, gl int) {
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

func (gs *GenomeSetHandler) geneTransposition(pgene float64) {
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

func (gs *GenomeSetHandler) onePointRecombination(p1p float64) {
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

func (gs *GenomeSetHandler) twoPointRecombination(p2p float64) {
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

func (gs *GenomeSetHandler) geneRecombination(pr float64) {
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

func (g *GenomeSetHandler) generate(iterate int, best *genomeset.GenomeSet) {
	saveCopy := best.Dup()
	g.replication() // Section 3.3.1, book page 75
	if g.Op.Conf.PMutate <= 0 || g.Op.Conf.PMutate >= 1 {
		config.ErrorF("pmutate 必须是 0 至 1 之间")
	} else {
		g.mutation(g.Op.Conf.PMutate) // Section 3.3.2, book page 77
	}

	if g.Op.Conf.Pis <= 0 || g.Op.Conf.Pis >= 1 {
		config.ErrorF("pis must be between 0 and 1")
	} else {
		g.isTransposition(g.Op.Conf.Pis, g.Op.Conf.Glis)
	}
	if g.Op.Conf.Pris <= 0 || g.Op.Conf.Pris >= 1 {
		config.ErrorF("pris must be between 0 and 1")
	} else {
		g.risTransposition(g.Op.Conf.Pris, g.Op.Conf.Glris)
	}
	if g.Op.Conf.PGene <= 0 || g.Op.Conf.PGene >= 1 {
		config.ErrorF("pgene must be between 0 and 1")
	} else {
		g.geneTransposition(g.Op.Conf.PGene)
	}
	if g.Op.Conf.P1p <= 0 || g.Op.Conf.P1p >= 1 {
		config.ErrorF("p1p must be between 0 and 1")
	} else {
		g.onePointRecombination(g.Op.Conf.P1p)
	}
	if g.Op.Conf.P2p <= 0 || g.Op.Conf.P2p >= 1 {
		config.ErrorF("p2p must be between 0 and 1")
	} else {
		g.twoPointRecombination(g.Op.Conf.P2p)
	}
	if g.Op.Conf.Pr <= 0 || g.Op.Conf.Pr >= 1 {
		config.ErrorF("pr must be between 0 and 1")
	} else {
		g.geneRecombination(g.Op.Conf.Pr)
	}
	// Now that replication is done, restore the best genome (aka "elitism")
	g.GenomeSets[0] = saveCopy

	for index, v := range g.GenomeSets {
		v.Iterate = iterate
		v.Index = index
	}
}

func (g *GenomeSetHandler) Evolve() *Record {
	best, accomplished := g.Op.Perf2(0, g.GenomeSets) // Preserve the best genome
	if accomplished {
		return g.makeRecord(best)
	}

	for i := 1; i <= g.Op.Conf.Iteration; i++ {
		g.generate(i, best)

		best, accomplished = g.Op.Perf2(i, g.GenomeSets) // Preserve the best genome
		if accomplished {
			return g.makeRecord(best)
		}
	}

	return g.makeRecord(best)
}

func (g *GenomeSetHandler) RunOnce() *Record {
	g.Op.GenomeSetF(g.GenomeSet)

	return g.makeRecord(g.GenomeSet)
}
