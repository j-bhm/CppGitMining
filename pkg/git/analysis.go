package git

import (
	"time"
	"math"
	
	"github.com/j-bhm/CppGitMining/pkg/allen"
	"github.com/j-bhm/CppGitMining/pkg/util"
	
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Analyse the git repository in the given path.
// Outputs a map with the following fields:
//   ContributorCount      int
//   ContributorEntropy    float64
//   CommitCount           int
//   BranchCount           int
//   FileCount             int
//   Lifetime              time.Duration
//   GitSize               float64
//   GitComplexity         float64
//   AvgContributorCommits float64
//   AvgBranchCommits      float64
func RunGitAnalysis(path string, opts util.Options) (map[string]interface{}, error){
    // open the repository
    util.PrintDebug("opening repository", opts)
    repo, err := git.PlainOpen(path)
    
    if err != nil{
        return nil, err
    }
    
    // analyse the repository
    result, err := AnalyseRepo(repo, opts)
    
    if err != nil{
        return nil, err
    }
    
    // return the result
    return result, nil
}

// Analyse the git repository repo.
// Outputs a map with the following fields:
//   ContributorCount      int
//   ContributorEntropy    float64
//   CommitCount           int
//   BranchCount           int
//   FileCount             int
//   Lifetime              time.Duration
//   GitSize               float64
//   GitComplexity         float64
//   AvgContributorCommits float64
//   AvgBranchCommits      float64
func AnalyseRepo(repo *git.Repository, opts util.Options) (map[string]interface{}, error){
    // get iterater over commits
    util.PrintDebug("analysing repository", opts)
    commitIter, err := repo.CommitObjects()
    
    if err != nil{
        return nil, err
    }
    
    // get head reference
    headRef, err := repo.Head()
    
    if err != nil{
        return nil, err
    }
    
    // get head commit
    headCommit, err := repo.CommitObject(headRef.Hash())
    
    if err != nil{
        return nil, err
    }
    
    // get iterator over files
    fileIter, err := headCommit.Files()
    
    if err != nil{
        return nil, err
    }
    
    // count files
    fileCount := 0
    fileIter.ForEach(func(file *object.File) error{
        fileCount += 1
        return nil
    })
    
    // define commit analysis variables
    var firstCommit time.Time
    var lastCommit time.Time
    commitCount := 0
    branchCount := 1
    parents := make(map[plumbing.Hash]bool) // tracks parents to recognise branch points
    authors := make(map[string]int) // maps authors to their number of commits
    nodes := make(map[plumbing.Hash]*allen.GraphNode) // map of nodes for the commit graph
    
    // iterate over commits
    first := true
    err = commitIter.ForEach(func(commit *object.Commit) error{
        // set time variables
        if first{
            firstCommit = commit.Author.When
            lastCommit = commit.Author.When
            first = false
        } else{
            // compare and update times
            date := commit.Author.When
            if date.After(lastCommit){
                lastCommit = date
            } else if date.Before(firstCommit){
                firstCommit = date
            }
        }
        
        // update counters on non-merge commits
        noMerge := commit.NumParents() <= 1
        if(noMerge){
            // increment author commits
            authors[commit.Author.Name] += 1
            
            // increment commit counter
            commitCount += 1
        }
        
        // get/create node for commit graph
        node := nodes[commit.Hash]
        
        if node == nil{
            node = new(allen.GraphNode)
            nodes[commit.Hash] = node
        }
        
        // iterate over commit parents
        parentIter := commit.Parents()
        err = parentIter.ForEach(func(parent *object.Commit) error{
            // get/create parent node
            parentNode := nodes[parent.Hash]
            
            if parentNode == nil{
                parentNode = new(allen.GraphNode)
                nodes[parent.Hash] = parentNode
            }
            
            // add edge to commit graph
            node.OutEdges = append(node.OutEdges, allen.GraphEdge{parentNode, 1.0})
            parentNode.InEdges = append(parentNode.InEdges, allen.GraphEdge{node, 1.0})
            
            // update parent tracker on non-merge commits
            if(noMerge){
                // test if parent has not been set yet
                if !parents[parent.ID()]{
                    // set parent as seen
                    parents[parent.ID()] = true
                } else{
                    // increase branch counter
                    branchCount += 1
                }
            }
            
            return nil
        })
        
        return err
    })
    
    if err != nil{
        return nil, err
    }
    
    // calculate commit graph complexity and size
    var commitGraph allen.Graph
    for _, v := range(nodes){
        commitGraph = append(commitGraph, v)
    }
    
    commitSize := allen.EstGraphSize(commitGraph)
    commitComplexity := allen.EstGraphComplexity(commitGraph)
    
    // calculate author based measures
    authorCount := 0
    authorEntropy := 0.0
    for _, v := range(authors){
        authorCount += 1
        p := float64(v) / float64(commitCount)
        authorEntropy -= p * math.Log2(p)
    }
    
    // set result values
    result := make(map[string]interface{})
    result["ContributorCount"] = authorCount
    result["ContributorEntropy"] = authorEntropy
    result["CommitCount"] = commitCount
    result["BranchCount"] = branchCount
    result["FileCount"] = fileCount
    result["Lifetime"] = lastCommit.Sub(firstCommit).Hours()
    result["GitSize"] = commitSize
    result["GitComplexity"] = commitComplexity
    result["AvgContributorCommits"] = float64(commitCount) / float64(authorCount)
    result["AvgBranchCommits"] = float64(commitCount) / float64(branchCount)
    
    // return result
    return result, nil
}
