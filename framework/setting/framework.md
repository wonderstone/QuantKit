::: mermaid
classDiagram

Framework --* Resource : composition
RTFramework --* Framework : composition
ReplayFramework --* Framework : composition
ResourceStruct --* Setting : composition
ReplayFramework *-- Setting : composition

ReplayFramework ..> ResourceStruct : dependency

ResourceStruct ..|> Resource : Realization

style Framework fill:#f91,stroke:#393,stroke-width:9px


namespace handler {
    class Framework {
    <<Interface>>
    Resource

	CurrTime() *time.Time
	CurrDate() time.Time
	TrainInfo() *model.TrainInfo, bool
	ConfigStrategyFromFile(file ...string) config.StrategyConfig
	Evaluate(values model.InputValues) model.OutputValues
    }

   class  Resource {
	<<Interface>>

	Dir() *config.Path
	Config() *config.Runtime
	Base() Basic
	Contract() Contract
	Quote() Quote
	Account() Account
	Creator() StrategyCreator
	Framework() Framework
    }

}


namespace setting {

    class RTFramework {
    <<Interface>>
    handler.Framework

	Init(resource Resource, option ...WithResource) error
	SetGenome(genome *genome.Genome)
	SetGenomeSet(genomeSet *genomeset.GenomeSet)
	SetStrategy(strategy handler.Strategy)
	Matcher() matcher.Matcher
	SubscribeData()
	Run()
	IsFinished() bool
	GetPerformance() float64
    }


    class ReplayFramework {
    <<Interface>>
    handler.Framework

	Init(resource Resource, option ...WithResource) error
	SetGenome(genome *genome.Genome)
	SetGenomeSet(genomeSet *genomeset.GenomeSet)
	SetStrategy(strategy handler.Strategy)
	Matcher() matcher.Matcher
	SubscribeData()
	Run()
	IsFinished() bool
	GetPerformance() float64
    }

    class ResourceStruct {
    <<struct>>
	config *config.Runtime

	basic handler.Basic

	contract handler.Contract

	quote handler.Quote

	account handler.Account

	framework handler.Framework

	strategyCreator handler.StrategyCreator

	setting Setting
    }

    class Setting {
    <<struct>>

	account handler.AccountSetting
	framework ReplayFramework
    }

    class Runner {
    <<Interface>>
	handler.Global

	Init(creator ...WithResource) error
	Start() error
    }

}


:::