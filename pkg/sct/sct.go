package sct

import (
    "errors"
    "encoding/json"
    "path/filepath"
	"os"
	"os/exec"
	
	"github.com/j-bhm/CppGitMining/pkg/allen"
	"github.com/j-bhm/CppGitMining/pkg/util"
)

// struct representing a node in a sct result
type SctNode struct{
    Id string
    Label string
    Value int
}

// struct representing an edge in a sct result
type SctEdge struct{
    Directed bool
    End string
    Id string
    Start string
    Weight float64
}

// struct representing the result of a sct run
type SctJson struct{
    Edges []SctEdge
    Nodes []SctNode
}

// directory for the results of the sct
const SctOutDir string = util.OutDir + "/sct"

// Run the sct and corresponding analysis
// on the repository specified in path
// and returns a map with the following fields:
//   SumScd        float64
//   MaxScd        float64
//   AvgScd        float64
//   SizeSct       float64
//   ComplexitySct float64
func RunSctAnalysis(path string, opts util.Options) (map[string]interface{}, error){
    // directory for sct output
    outputDir := SctOutDir + "/" + filepath.Base(path)
    
    // check if a sct output already exists
    _, err := os.Stat(outputDir + "/0/results.json")
    
    if err == nil{
        // check if old output should be ignored
        if opts.ForceSct {
            // remove old output
            err := os.RemoveAll(outputDir)
                
            if err != nil{
                return nil, err
            }
        } else{
        // try parsing old output
        util.PrintDebug("parsing old sct result", opts)
        sctJson, err := ParseSctOutput(outputDir + "/0")
        
        if err == nil{
                // analyse output
                util.PrintDebug("using old result for analysis", opts)
                sctResult := AnalyseSctOutput(sctJson)
                
                // return result
                return sctResult, nil
            } else{
                // delete old output
                util.PrintDebug("parsing failed: " + err.Error(), opts)
                err := os.RemoveAll(outputDir)
                
                if err != nil{
                    return nil, err
                }
            }
        }
    }
    
    // create directory for sct output
    err = os.MkdirAll(outputDir, 0750)
    
    if err != nil{
        return nil, err
    }

    // run sct on path
    util.PrintDebug("running static coupling tool", opts)
    err = RunSct(path, outputDir, opts)
    
    if err != nil{
        os.RemoveAll(outputDir)
        return nil, err
    }
    
    // parse sct output
    util.PrintDebug("parsing sct output", opts)
    sctJson, err := ParseSctOutput(outputDir + "/0")
    
    if err != nil{
        return nil, err
    }
    
    // analyse sct output
    util.PrintDebug("analysing sct output", opts)
    sctResult := AnalyseSctOutput(sctJson)
    
    // return result
    return sctResult, nil
}

// Analyse a sct json and return the result
// with the fields:
//   SumScd        float64
//   MaxScd        float64
//   AvgScd        float64
//   SizeSct       float64
//   ComplexitySct float64
func AnalyseSctOutput(sctJson *SctJson) map[string]interface{}{
    // create result
    result := make(map[string]interface{})
    
    // calculate static coupling degrees
    graph := ConvertSctToGraph(sctJson)
    scds := allen.SumWeights(graph)
    
    // calculate static coupling metrics
    sum := 0.0
    maximum := 0.0
    for _, v := range(scds){
        if v > maximum{
            maximum = v
        }
        
        sum += v
    }
    
    // save results
    result["SumScd"] = sum
    result["MaxScd"] = maximum
    result["AvgScd"] = sum / float64(len(scds))
    
    // calculate alan metric
    result["SizeSct"] = allen.EstGraphSize(graph)
    result["ComplexitySct"] = allen.EstGraphComplexity(graph)
    
    return result
}

// Run the StaticCouplingTool on the
// repository specified in path and
// save the results in the directory outDir.
func RunSct(path, outDir string, opts util.Options) error{
    // create command executing sct
    cmdArgs := []string{"-m", "-l", "cpp", "-p", path, "-o", outDir}
    cmd := exec.Command(opts.Sct, cmdArgs...)
    
    // run the command
    err := util.RunCmd(cmd, opts)
    
    if err != nil{
        return err
    }
    
    // return success
    return nil
}

// Convert the json result from the sct
// into the graph representation used
// for analysis.
func ConvertSctToGraph(json *SctJson) allen.Graph{

    // copy nodes
    graphMap := make(map[string]*allen.GraphNode)
    for _, node := range(json.Nodes){
        graphMap[node.Id] = new(allen.GraphNode)
    }

    // copy edges
    for _, edge := range(json.Edges){
        // get graph nodes
        startNode := graphMap[edge.Start]
        endNode := graphMap[edge.End]
        
        // add edge to graph
        startNode.OutEdges = append(startNode.OutEdges, allen.GraphEdge{endNode, edge.Weight})
        endNode.InEdges = append(endNode.InEdges, allen.GraphEdge{startNode, edge.Weight})
    }
    
    // create slice
    graph := make([]*allen.GraphNode, 0, len(graphMap))
    for _, node := range(graphMap){
        graph = append(graph, node)
    }

    // return result
    return graph
}

// Parse and return the results.json in the specified directory.
// Returns an error if the parsing fails.
func ParseSctOutput(dir string) (*SctJson, error) {
    // read data from result.json
    data, err := os.ReadFile(dir + "/results.json")
    
    if err != nil{
        return nil, err
    }

    // parse data as json
    sctJson := new(SctJson)
    err = json.Unmarshal(data, sctJson)
    
    if err != nil{
        return nil, err
    }
    
    // test for empty graph
    if len(sctJson.Nodes) == 0{
        return nil, errors.New("empty sct graph")
    }
    
    // return data
    return sctJson, nil
}
