package gep

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/wonderstone/QuantKit/config"

	"testing"

	"github.com/wonderstone/QuantKit/modelgene/gep/functions"
	"github.com/wonderstone/QuantKit/modelgene/gep/gene"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/grammars"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
)

// srTests is a random sample of inputs and outputs for the function "a^4 + a^3 + a^2 + a"
var srTests1 = []struct {
	in  []float64
	out float64
}{
	// {[]float64{0}, 0},
	{[]float64{2.81}, 95.2425},
	{[]float64{6}, 1554},
	{[]float64{7.043}, 2866.55},
	{[]float64{8}, 4680},
	{[]float64{10}, 11110},
	{[]float64{11.38}, 18386},
	{[]float64{12}, 22620},
	{[]float64{14}, 41370},
	{[]float64{15}, 54240},
	{[]float64{20}, 168420},
	{[]float64{100}, 101010100},
	{[]float64{-100}, 99009900},
}
var srTests2 = []struct {
	in  []float64
	out float64
}{
	// {[]float64{0}, 0},
	{[]float64{9.36, 3.0}, 3.12},
	{[]float64{6.93, 2.31}, 3.0},
	{[]float64{44.33, 22.165}, 2.0},
	{[]float64{666.5344, 8.96}, 74.39},
	{[]float64{852.3392, 10.88}, 78.34},
	{[]float64{1016.234, 11.38}, 89.3},
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func validateFunc(g *genome.Genome) float64 {
	result := 0.0
	for _, n := range srTests1 {
		r := g.EvalMath(n.in)
		// fmt.Printf("r=%v, n.in=%v, n.out=%v, g=%v\n", r, n.in, n.out, g)
		if math.IsInf(r, 0) {
			return 0.0
		}
		fitness := math.Abs(r - n.out)
		fitness = 1000.0 / (1.0 + fitness) // fitness is normalized and max value is 1000
		// fmt.Printf("r=%v, n.in=%v, n.out=%v, fitness=%v, g=%v\n", r, n.in, n.out, fitness, g)
		result += fitness
	}
	return result / float64(len(srTests1))
}

// 测试GEP符号+-*/
func TestGEPSymbol(t *testing.T) {
	funcs := []gene.FuncWeight{
		{Symbol: "+", Weight: 1},
		{Symbol: "-", Weight: 1},
		{Symbol: "*", Weight: 1},
		{Symbol: "/", Weight: 1},
	}
	numIn := len(srTests1[0].in)
	e := model.New(funcs, functions.Float64, 50, 4, 2, numIn, 0, "+", validateFunc)
	s := e.Evolve(1000, 50000, 0.8, 0.5, 3, 0.5, 3, 0.5, 0.01, 0.01, 0.01)

	// Write out the Go source code for the solution.
	gr, err := grammars.LoadGoMathGrammar()
	if err != nil {
		log.Printf("unable to load grammar: %v", err)
	}
	fmt.Printf("\n// gepModel is auto-generated Go source code for the\n")
	fmt.Printf("// (a^4 + a^3 + a^2 + a) solution karva expression:\n// %q, score=%v\n", s, validateFunc(s))
	s.Write(os.Stdout, gr)
	fmt.Println(s.Genes, validateFunc(s))
	// gene1 := gene.New("/.d0.d1.-.d0.d0.d0.d0.d1", functions.Float64)
	// gene2 := gene.New("-.-.-.d0.d1.d0.d0.d1.d0", functions.Float64)
	// genome1 := genome.New([]*gene.Gene{gene1, gene2}, "+")
	// fmt.Println(validateFunc(genome1))

}

func TestGEP(t *testing.T) {
	m, _ := config.NewModelConfig("/Users/alexxiong/GolandProjects/quant/strategy/S20230824-090000-000/.vqt/train/T20230824-110000-000/input/model.yaml")

	model.NewHandler(
		config.ModelTypeGenome,
		model.WithModelConfig(*m.Gep),
		model.WithKarvaExpressionFile("/Users/alexxiong/GolandProjects/quant/strategy/S20230824-090000-000/.vqt/bt/B20230824-110000-000/input/expression.yaml"),
		model.WithGenome(
			func(g *genome.Genome) {
				fmt.Println(
					g.EvalMath(
						[]float64{
							7.68, 7.27, 7.33, 7.3, 24222613, 176749046, 7.32, 7.31, 7.31, 7.31,
							// 15.3, 14.31, 15.01, 14.56, 30688983, 445475572, math.NaN(), math.NaN(), math.NaN(),
						},
					),
				)
				gr, _ := grammars.LoadGoMathGrammar()

				fmt.Println(g.WriteExpsStr(gr, []string{"Open", "Close", "High", "Low", "MA3", "MA5", "MA8", "MA10"}))
			},
		),

		model.WithNumTerminal(8),
	).RunOnce()

}
