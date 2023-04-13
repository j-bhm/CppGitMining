package gct

import(
    "errors"
    "encoding/json"
    "path/filepath"
    "os"
    "os/exec"

    "github.com/j-bhm/CppGitMining/pkg/allen"
    "github.com/j-bhm/CppGitMining/pkg/util"
)

// struct representing a node in a gct result
type GctNode struct{
    Id string
}

// struct representing an edge in a gct result
type GctEdge struct{
    Id string
    Start string
    End string
    Weight float64
    Directed bool
}

// struct representing the result of a gct run
type GctJson struct{
    Nodes []GctNode
    Edges []GctEdge
}

// directory for the results of the gct
const GctOutDir string = util.OutDir + "/gct"

// Run the gct and corresponding analysis on the repository in path
// and return a map with the following fields:
//   SumGcd        float64
//   MaxGcd        float64
//   AvgGcd        float64
//   SizeGct       float64
//   ComplexityGct float64
func RunGctAnalysis(path string, opts util.Options) (map[string]interface{}, error){
    // path for gct output
    outputDir := GctOutDir + "/" + filepath.Base(path)
    
    // check if a gct output already exists
    _, err := os.Stat(outputDir + "/result.json")
    
    if err == nil{
        // check if old output should be ignored
        if opts.ForceGct {
            // remove old output
            err := os.RemoveAll(outputDir)
            
            if err != nil{
                return nil, err
            }
        } else{
            // try parsing old output
            util.PrintDebug("parsing old gct result", opts)
            gctJson, err := ParseGctOutput(outputDir)
            
            if err == nil{
                // analyse output
                util.PrintDebug("using old result for analysis", opts)
                gctResult := AnalyseGctOutput(gctJson)
                
                // return result
                return gctResult, nil
            } else{
                // delete old output
                util.PrintDebug("parsing failed: "+  err.Error(), opts)
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

    // run the gct on path
    util.PrintDebug("running git coupling tool", opts)
    err = RunGct(path, outputDir, opts)
    
    if err != nil{
        os.RemoveAll(outputDir)
        return nil, err
    }
    
    // parse the gct output
    util.PrintDebug("parsing gct output", opts)
    gctJson, err := ParseGctOutput(outputDir)
    
    if err != nil{
        return nil, err
    }
    
    // analyse gct output
    util.PrintDebug("analysing gct output", opts)
    gctResult := AnalyseGctOutput(gctJson)
    
    // return result
    return gctResult, nil
}

// Analyse a gct json and return the result
// with the following fields:
//   SumGcd        float64
//   MaxGcd        float64
//   AvgGcd        float64
//   SizeGct       float64
//   ComplexityGct float64
func AnalyseGctOutput(gctJson *GctJson) map[string]interface{}{
     // result variable
     result := make(map[string]interface{})
    
    // calculate git coupling degrees
    graph := ConvertGctToGraph(gctJson)
    gcds := allen.SumWeights(graph)
    
    // calculate git coupling metrics
    sum := 0.0
    maximum := 0.0
    for _, v := range(gcds){
        if v > maximum{
            maximum = v
        }
        
        sum += v
    }
    
    // save results
    result["SumGcd"] = sum
    result["MaxGcd"] = maximum
    result["AvgGcd"] = sum / float64(len(gcds))
    
    // calculate alan metric
    result["SizeGct"] = allen.EstGraphSize(graph)
    result["ComplexityGct"] = allen.EstGraphComplexity(graph)
    
    return result
}

// Run the GitCouplingTool on the
// repository specified in path and
// save the result in the directory outDir.
func RunGct(path, outDir string, opts util.Options) error {
    // create command executing sct
    cmdArgs := []string{path, "-r", "-c", "1", "--file-type", ".c", "--file-type", ".cpp", "--file-type", ".h", "--file-type", ".hpp", "-f", "JSON", "-o", outDir + "/result.json"}
    cmd := exec.Command(opts.Gct, cmdArgs...)
    
    // run the command
    err := util.RunCmd(cmd, opts)
    
    if err != nil{
        return err
    }
    
    // return success
    return nil
}

// Convert the json result from the gct
// into the graph representation used
// for analysis.
func ConvertGctToGraph(gct *GctJson) allen.Graph{

    // copy nodes
    graphMap := make(map[string]*allen.GraphNode)
    for _, node := range(gct.Nodes){
        graphMap[node.Id] = new(allen.GraphNode)
    }

    // copy edges
    for _, edge := range(gct.Edges){
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

// Parse and return a gct output
// from the results.json in the specified directory.
// Returns an error if the parsing fails.
func ParseGctOutput(dir string) (*GctJson, error) {
    // read data from result.json
    data, err := os.ReadFile(dir + "/result.json")
    
    if err != nil{
        return nil, err
    }

    // parse data as json
    gctJson := new(GctJson)
    err = json.Unmarshal(data, gctJson)
    
    if err != nil{
        return nil, err
    }
    
    // test for empty graph
    if len(gctJson.Nodes) == 0{
        return nil, errors.New("empty gct graph")
    }
    
    // return data
    return gctJson, nil
}
