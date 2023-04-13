package allen

import(
    "math"
)

// edge of a graph
type GraphEdge struct{
    Node *GraphNode
    Weight float64
}

// node of a graph
type GraphNode struct{
    InEdges []GraphEdge
    OutEdges []GraphEdge
}

// graph type
type Graph []*GraphNode

// Sum weights of all edges for every node.
func SumWeights(graph Graph) []float64{
    result := make([]float64, len(graph))
    
    // loop over nodes
    for i := range(graph){
        result[i] = 0
        
        // sum in edges
        for _, edge := range(graph[i].InEdges){
            result[i] += edge.Weight
        }
        
        // sum out edges
        for _, edge := range(graph[i].OutEdges){
            result[i] += edge.Weight
        }
    }
        
    return result
}

// Caluclate the edges-only graph of the
// given graph.
func EdgeOnlyGraph(graph Graph) (result Graph){
    // loop over all nodes
    for _, node := range(graph){
        // skip nodes without edges
        if len(node.OutEdges) == 0 && len(node.InEdges) == 0{
            continue
        }
        
        // otherwise add node to the result
        result = append(result, node)
    }
    return
}

// Calculate the estimated size of the given graph.
func EstGraphSize(graph []*GraphNode) float64{
    var zeroNodes float64 = 1
    var nodeCount float64 = float64(len(graph) + 1)
    var result float64 = 0.0
    
    // loop over all nodes
    for _, node := range(graph){
        // check if node is decoupled from the graph,
        // its label is zero
        if len(node.InEdges) == 0 && len(node.OutEdges) == 0{
            zeroNodes += 1
            continue
        }
        
        // check if node is only connected to one other node
        // its label appears twice
        if len(node.InEdges) == 0 && len(node.OutEdges) == 1{
            otherNode := node.OutEdges[0].Node
            if len(otherNode.InEdges) == 1 && len(otherNode.OutEdges) == 0{
                result -= math.Log2(2/nodeCount)
                continue
            }
        }
        
        if len(node.InEdges) == 1 && len(node.OutEdges) == 0{
            otherNode := node.InEdges[0].Node
            if len(otherNode.InEdges) == 0 && len(otherNode.OutEdges) == 1{
                result -= math.Log2(2/nodeCount)
                continue
            }
        }
        
        if len(node.InEdges) == 1 && len(node.OutEdges) == 1{
            inNode := node.InEdges[0].Node
            outNode := node.OutEdges[0].Node
            
            if inNode == outNode{
                if len(inNode.InEdges) == 1 && len(inNode.OutEdges) == 1{
                    result -= math.Log2(2/nodeCount)
                    continue
                }
            }
        }
        
        // otherwise its label is distinct
        result -= math.Log2(1/nodeCount)
    }
    
    // add zero labels excluding the environment node
    result -= (zeroNodes - 1) * math.Log2(zeroNodes/nodeCount)
    
    return result
}

// Calculate the estimated size of the
// node subsystem graph for node i.
func EstIGraphSize(i int, graph Graph) float64{
    node := graph[i]
    var nodeCount float64 = float64(len(graph) + 1)
    var result float64 = 0.0
    
    // get all nodes connected to node i
    connectedNodes := float64(len(node.InEdges))
    for _, outE := range(node.OutEdges){
        isInE := false
        for _, inE := range(node.InEdges){
            if(inE.Node == outE.Node){
                isInE = true
                break
            }
        }
        
        if !isInE{
            connectedNodes += 1
        }
    }
    
    // add labels for node i and connected nodes
    if connectedNodes == 1{
        // both nodes have the same label
        result -= 2 * math.Log2(2/nodeCount)
    } else{
        // all labels are distinct
        result -= (connectedNodes + 1) * math.Log2(1/nodeCount)
    }
    
    // add zero labels excluding the environment node
    zeroNodes := nodeCount - connectedNodes - 1
    result -= (zeroNodes - 1) * math.Log2(zeroNodes/nodeCount)
    
    return result
}

// Calculate the estimated complexity of the
// given graph.
func EstGraphComplexity(graph Graph) float64{
    // get the edges-only graph
    edgeOnlyGraph := EdgeOnlyGraph(graph)
    
    // sum all estimated subsystem graph sizes
    var result float64 = 0.0
    for i := range(edgeOnlyGraph){
        result += EstIGraphSize(i, edgeOnlyGraph)
    }
    
    // subtract the estimated edges-only graph size
    result -= EstGraphSize(edgeOnlyGraph)
    
    return result
}
