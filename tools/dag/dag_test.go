package dag

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"testing"
)

func TestDAG(t *testing.T) {
	// 创建一个有向图
	dg := simple.NewDirectedGraph()

	// 添加节点
	node1 := dg.NewNode()
	dg.AddNode(node1)
	node2 := dg.NewNode()
	dg.AddNode(node2)
	node3 := dg.NewNode()
	dg.AddNode(node3)

	// 添加边
	dg.SetEdge(dg.NewEdge(node1, node2))
	dg.SetEdge(dg.NewEdge(node2, node3))

	// 检查是否是有向无环图
	if IsDAG(dg) {
		fmt.Println("The graph is a DAG.")
	} else {
		fmt.Println("The graph is not a DAG.")
	}

	// 进行拓扑排序
	sorted, err := topo.Sort(dg)
	if err != nil {
		fmt.Println("Cannot do a topo sort:", err)
		return
	}

	// 输出排序结果
	for _, node := range sorted {
		fmt.Println(node.ID())
	}

	// 添加一条创建环的边
	dg.SetEdge(dg.NewEdge(node3, node1))

	// 再次检查
	if IsDAG(dg) {
		fmt.Println("The graph is a DAG.")
	} else {
		fmt.Println("The graph is not a DAG.")
	}
}
