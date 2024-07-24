package model

import (
	"math"
	"math/rand"

	"github.com/wonderstone/QuantKit/modelgene/gep/gene"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/config"
)

type GenomeHandler struct {
	Op      *Op
	Genomes []*genome.Genome
	Genome  *genome.Genome
	Funcs   []gene.FuncWeight
}

func (g *GenomeHandler) Init(option ...WithOption) error {
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

	if g.Op.GenomeF != nil {
		g.Funcs = functions
		genes := make([]*gene.Gene, g.Op.Conf.NumGenesPerGenome)
		for j := range genes {
			genes[j] = gene.New(g.Op.Record.Gep.KES[0][j], g.Op.FuncType)
		}
		g.Genome = genome.New(genes, g.Op.Conf.LinkFunc)
	} else if g.Op.Perf != nil {
		g.Genomes = make([]*genome.Genome, g.Op.Conf.NumGenomes)
		g.Funcs = functions
		n := maxArity(functions, g.Op.FuncType)
		tailSize := g.Op.Conf.HeadSize*(n-1) + 1
		for i := range g.Genomes {
			genes := make([]*gene.Gene, g.Op.Conf.NumGenesPerGenome)

			// for j := range genes {
			// 	genes[j] = gene.New(g.Op.Record.Gep.KES[0][j], g.Op.FuncType)
			// }
			for j := range genes {
				genes[j] = gene.RandomNew(
					g.Op.Conf.HeadSize, tailSize, g.Op.NumTerminal, g.Op.Conf.NumConstants, functions, g.Op.FuncType,
				)
			}
			g.Genomes[i] = genome.New(genes, g.Op.Conf.LinkFunc)
			g.Genomes[i].Index = i
			g.Genomes[i].Iterate = 0
			// fmt.Println("Genome: ", g.Genomes[i].StringSlice())
		}
	} else {
		config.ErrorF("未指定Genome模式下的筛选函数 或者 评估函数")
	}

	return nil
}

func (g *GenomeHandler) makeRecord(gene *genome.Genome) *Record {
	// 将karva表达式写入文件
	return &Record{
		Gep: GepRecord{
			Mode:  "Genome",
			Score: gene.Score,
			KES:   [][]string{gene.StringSlice()},
		},
	}
}

func (g *GenomeHandler) replication() {
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

func (g *GenomeHandler) mutation(pm float64) {
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

func (g *GenomeHandler) isTransposition(pis float64, gl int) {
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

func (g *GenomeHandler) risTransposition(pris float64, gl int) {
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

func (g *GenomeHandler) geneTransposition(pgene float64) {
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

func (g *GenomeHandler) onePointRecombination(p1p float64) {
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

func (g *GenomeHandler) twoPointRecombination(p2p float64) {
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

func (g *GenomeHandler) geneRecombination(pr float64) {
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

func (g *GenomeHandler) generate(iterate int, bestGenome *genome.Genome) {
	saveCopy := bestGenome.Dup()
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
	g.Genomes[0] = saveCopy

	for index, v := range g.Genomes {
		v.Iterate = iterate
		v.Index = index
	}
}

func (g *GenomeHandler) Evolve() *Record {
	best, accomplished := g.Op.Perf(0, g.Genomes) // Preserve the best genome
	if accomplished {
		return g.makeRecord(best)
	}

	for i := 1; i <= g.Op.Conf.Iteration; i++ {
		g.generate(i, best)

		best, accomplished = g.Op.Perf(i, g.Genomes) // Preserve the best genome
		if accomplished {
			return g.makeRecord(best)
		}
	}

	return g.makeRecord(best)
}

func (g *GenomeHandler) RunOnce() *Record {
	g.Op.GenomeF(g.Genome)

	return g.makeRecord(g.Genome)
}
