package dag

import (
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

// DAG is a directed acyclic graph.
func IsDAG(g *simple.DirectedGraph) bool {
	_, err := topo.Sort(g)
	return err == nil
}
