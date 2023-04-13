package main

import "encoding/json"
import "flag"
import "fmt"
import "os"
import "path/filepath"

import "github.com/j-bhm/CppGitMining/pkg/util"
import "github.com/j-bhm/CppGitMining/pkg/sct"
import "github.com/j-bhm/CppGitMining/pkg/gct"
import "github.com/j-bhm/CppGitMining/pkg/git"

// struct used to export the analysis results
type result struct{
    Git map[string]interface{}
    Sct map[string]interface{}
    Gct map[string]interface{}
}

func main(){
    // define flags
    var verbFlag = flag.Int("v", 1, "verbosity:\n0: only print error messages\n1: print status messages\n2: print debug messages\n3: pass on output of shell commands")
    var sctFlag = flag.String("sct", "StaticCouplingTool", "command to run the StaticCouplingTool")
    var gctFlag = flag.String("gct", "GitCouplingTool", "path to the GitCouplingTool jar file")
    var skipGitFlag = flag.Bool("skip-git", false, "skip extraction of git metrics")
    var skipSctFlag = flag.Bool("skip-sct", false, "skip analysis based on the static coupling tool")
    var skipGctFlag = flag.Bool("skip-gct", false, "skip analysis based on the git coupling tool")
    var skipBuildFlag = flag.Bool("skip-build", false, "skip build process and only clone the repositories")
    var forceGitFlag = flag.Bool("force-git", false, "ignore old saves and reload every git")
    var forceSctFlag = flag.Bool("force-sct", false, "ignore old sct outputs and rerun analysis")
    var forceGctFlag = flag.Bool("force-gct", false, "ignore old gct outputs and rerun analysis")
    var outputFlag = flag.String("o", "./result.json", "file to save output in")
 
    // parse flags
    flag.Parse()
    
    // check non-flag arguments
    numArgs := flag.NArg()
    
    if numArgs == 0{
        fmt.Println("missing argument: path to input file is required")
        fmt.Println("usage: cgm [options] <path>")
        return
    } else if numArgs > 1{
        fmt.Println("too many arguments:")
        fmt.Println(flag.Args())
        fmt.Println("only the input file path is required")
        fmt.Println("usage: cgm [options] <path>")
        return
    }
    
    // set options
    var opts util.Options
    opts.Verbosity = *verbFlag
    opts.Sct = *sctFlag
    opts.Gct = *gctFlag
    opts.SkipBuild = *skipBuildFlag
    opts.ForceGit = *forceGitFlag
    opts.ForceSct = *forceSctFlag
    opts.ForceGct = *forceGctFlag
    
    // parse input file
    inputPath := flag.Arg(0)
    urls, commands, err := util.ParseInput(inputPath)
    
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    
    // clone and build repositories
    util.PrintStatus("loading repositories:", opts)
    repos := git.LoadRepos(urls, commands, opts)
    
    // run analyses
    util.PrintStatus("analysing repositories:", opts)
    output := make(map[string]result)
    for i := range(repos){
        repo := repos[i]
        var res result
        util.PrintStatus(fmt.Sprintf("[%d/%d] %s", i + 1, len(repos), repo), opts)
        
        // run git analysis
        if !*skipGitFlag{
            util.PrintStatus(fmt.Sprintf("[%d/%d] running git analysis", i + 1, len(repos)), opts)
            gitResult, err := git.RunGitAnalysis(repo, opts)
            
            if err != nil{
                util.PrintError(err.Error(), opts)
                continue
            }
            
            res.Git = gitResult
        }
        
        // run sct analysis
        if !*skipSctFlag{
            util.PrintStatus(fmt.Sprintf("[%d/%d] running static coupling analysis", i + 1, len(repos)), opts)
            sctResult, err := sct.RunSctAnalysis(repo, opts)
            
            if err != nil{
                util.PrintError(err.Error(), opts)
                continue
            }
            
            res.Sct = sctResult
        }
        
        // run gct analysis
        if !*skipGctFlag{
            util.PrintStatus(fmt.Sprintf("[%d/%d] running git coupling analysis", i + 1, len(repos)), opts)
            gctResult, err := gct.RunGctAnalysis(repo, opts)
            
            if err != nil{
                util.PrintError(err.Error(), opts)
                continue
            }
            
            res.Gct = gctResult
        }
        
        // add result to output
        gitName := filepath.Base(repo)
        output[gitName] = res
    }
    
    // create json data of output
    util.PrintStatus("saving output: ", opts)
    data, err := json.MarshalIndent(output, "", "    ")
    
    if err != nil{
        fmt.Println(err.Error())
        return
    }
    
    // create output file
    file, err := os.Create(*outputFlag)
    
    if err != nil{
        fmt.Println(err.Error())
        return
    }
    
    // write data to output file
    _, err = file.Write(data)
    
    if err != nil{
        fmt.Println(err.Error())
        return
    }
    
    // close file
    err = file.Close()
    
    if err != nil{
        fmt.Println(err.Error())
    }
}
