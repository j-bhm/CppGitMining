package git

import (
    "fmt"
    "os"
    "os/exec"
    "path"
    "strings"
    
    "github.com/j-bhm/CppGitMining/pkg/util"
    
    "github.com/go-git/go-git/v5"
)

// directory with git repositories
const GitDir = util.OutDir + "/gits"

// Clone all repositories given in urls and
// build each one with the corresponding command
// in commands.
// Returns the paths to the cloned repositories.
func LoadRepos(urls []string, commands []string, opts util.Options) (paths []string) {
    // test for correct input
    if len(urls) != len(commands){
        panic("loading repositories: unequal number of urls and build commands")
    }
    
    // loop over inputs
    for i := range(urls){
        // compute directory for the repository
        dir := GitDir + "/"
        dir += strings.TrimSuffix(path.Base(urls[i]), ".git")
        
        // test if the repository already exists
        util.PrintStatus(fmt.Sprintf("[%d/%d] loading repository: %s", i + 1, len(urls), urls[i]), opts)
        _, err := git.PlainOpen(dir)
        
        if err == nil {
            // check if repo should be reloaded
            if opts.ForceGit {
                // remove old repo
                err := os.RemoveAll(dir)
                
                if err != nil{
                    util.PrintError(err.Error(), opts)
                    continue
                }
            } else{
                // add dir to the results
                paths = append(paths, dir)
                util.PrintDebug("repository already exists", opts)
                continue
            }
        }
    
        // load the repository
        err = LoadRepo(urls[i], dir, commands[i], opts)
        
        if err != nil{
            util.PrintError(err.Error(), opts)
            os.RemoveAll(dir)
            continue
        }
        
        // add dir to the results
        paths = append(paths, dir)
    }
    
    // return all paths
    return paths
}



// Clone the repository from url into
// the directory dir and
// build it with the given command.
func LoadRepo(url, dir, command string, opts util.Options) error{ 
    // clone repository
    util.PrintDebug("cloning repository", opts)
    err := CloneRepo(url, dir)

    if err != nil{
        return err
    }
    
    // build repository
    if !opts.SkipBuild{
        util.PrintDebug("building repository:", opts)
        err = BuildRepo(dir, command, opts)
        
        if err != nil{
            return err
        }
    }
    
    // return
    return nil
}

// Clone the repository given
// by url into the directory dir.
func CloneRepo(url string, dir string) error {
    // clone repo from url
    _, err := git.PlainClone(dir, false, &git.CloneOptions{
        URL: url,
    })
    
    if err != nil{
        return err
    }
    
    // return success
    return nil
}

// Build the repository at the given path
// with the given command.
func BuildRepo(path string, command string, opts util.Options) error {
    // create build command
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = path
    
    // run the command
    err := util.RunCmd(cmd, opts)
    
    if err != nil{
        return err
    }
    
    // return success
    return nil
}
