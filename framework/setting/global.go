package setting

import (
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
)

type Setting struct {
	account handler.Accounts

	framework ReplayFramework
}

type Resource struct {
	config *config.Runtime

	basic handler.Basic

	contract handler.Contract

	quote handler.Quote

	account handler.Account

	framework handler.Framework

	strategyCreator handler.StrategyCreator

	setting Setting
}

type WithResource func(r *Resource)

func WithGenome(genome *genome.Genome) WithResource {
	return func(r *Resource) {
		r.setting.framework.SetGenome(genome)
	}
}

func WithGenomeSet(genomeset *genomeset.GenomeSet) WithResource {
	return func(r *Resource) {
		r.setting.framework.SetGenomeSet(genomeset)
	}
}

func WithRuntimeConfig(config *config.Runtime) WithResource {
	return func(r *Resource) {
		r.config = config
	}
}

func WithBaseHandler(basic handler.Basic) WithResource {
	return func(r *Resource) {
		r.basic = basic
	}
}

func WithAccountHandler(account handler.Account) WithResource {
	return func(r *Resource) {
		r.account = account
	}
}

func WithContractHandler(contract handler.Contract) WithResource {
	return func(r *Resource) {
		r.contract = contract
	}
}

func WithQuoteHandler(quote handler.Quote) WithResource {
	return func(r *Resource) {
		r.quote = quote
	}
}

func WithStrategyCreator(creator handler.StrategyCreator) WithResource {
	return func(r *Resource) {
		r.strategyCreator = creator
	}
}

func WithFramework(framework handler.Framework) WithResource {
	return func(r *Resource) {
		r.framework = framework
	}
}

func NewResource(options ...WithResource) *Resource {
	r := &Resource{}

	for _, option := range options {
		option(r)
	}

	return r
}

func (g Resource) Dir() *config.Path {
	return g.config.Path
}

func (g Resource) Config() *config.Runtime {
	return g.config
}

func (g Resource) Base() handler.Basic {
	return g.basic
}

func (g Resource) Contract() handler.Contract {
	return g.contract
}

func (g Resource) Quote() handler.Quote {
	return g.quote
}

func (g Resource) Creator() handler.StrategyCreator {
	return g.strategyCreator
}

func (g Resource) Account() handler.Account {
	return g.account
}

func (g Resource) Framework() handler.Framework {
	return g.framework
}

func (g *Resource) Set(options ...WithResource) {
	for _, option := range options {
		option(g)
	}
}
