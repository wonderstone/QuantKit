package gep

import (
	"github.com/wonderstone/QuantKit/modelgene/gep/functions"
	"github.com/wonderstone/QuantKit/modelgene/gep/gene"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/grammars"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/config"

	"github.com/jinzhu/copier"
)

type ModelGEP struct {
	Config config.Model

	Generation model.Generation
}

func Run(config config.Model, expectFitness float64, numTerminals int, validF genome.ScoringFunc, outputF func(g *genome.Genome, gr *grammars.Grammar)) error {
	var fw []gene.FuncWeight
	err := copier.Copy(&fw, config.Gep.Function)
	if err != nil {
		return err
	}

	switch config.Gep.Mode {
	case "Genome":
		m := model.New(fw, functions.Float64, config.Gep.NumGenomes, config.Gep.HeadSize, config.Gep.NumGenesPerGenome, numTerminals, config.Gep.NumConstants, config.Gep.LinkFunc, validF)
		s := m.Evolve(config.Gep.Iteration, expectFitness, config.Gep.PMutate, config.Gep.Pis, config.Gep.Glis, config.Gep.Pris, config.Gep.Glris, config.Gep.PGene, config.Gep.P1p, config.Gep.P2p, config.Gep.Pr)
		gr, err := grammars.LoadGoMathGrammar()
		if err != nil {
			return err
		}

		outputF(s, gr)
	}

	return nil
}
