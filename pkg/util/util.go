package util

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "os/exec"
)

// directory for all output/temp files and directories
const OutDir = ".mp"

type Options struct{
    Verbosity int  // option controlling the message printing
    Sct string     // command to execute the StaticCouplingTool
    Gct string     // command to execute the GitCouplingTool
    SkipBuild bool // skip the build process
    ForceGit bool  // reload gits ignoring old saves
    ForceSct bool  // rerun sct analysis ignoring old outputs
    ForceGct bool  // rerun gct analysis ignoring old outputs
}

// Print an error message.
func PrintError(msg string, opts Options){
    if opts.Verbosity >= 0 {
        fmt.Println(msg)
    }
}

// Print a status message.
func PrintStatus(msg string, opts Options){
    if opts.Verbosity >= 1 {
        fmt.Println(msg)
    }
}

// Print a debug message.
func PrintDebug(msg string, opts Options){
    if opts.Verbosity >= 2 {
        fmt.Println(msg)
    }
}

// Run the command cmd.
// If verbosity >= 3, passes the
// input and output on to standard io.
func RunCmd(cmd *exec.Cmd, opts Options) error {
    // set cmd outputs
    if opts.Verbosity >= 3 {
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stdout
        cmd.Stdin = os.Stdin
    }
    
    // run the command
    PrintDebug("executing: " + cmd.String(), opts)
    err := cmd.Run()
    
    if err != nil {
        return err
    }
    
    // return success
    return nil
}

// Parse input file given at path.
// Returns the repository urls
// and building commands or an error
// if parsing fails.
func ParseInput(path string) (urls []string, commands []string, err error) {
    // open file
    file, err := os.Open(path)
    
    if err != nil{
        return nil, nil, err
    }
    defer file.Close()
    
    // scan file for urls and commands
    scanner := bufio.NewScanner(file)
    for scanner.Scan(){
        // read repository link
        url := scanner.Text()
        urls = append(urls, url)
        
        // read command line
        ok := scanner.Scan()
        if !ok {
            return nil, nil, errors.New("parsing " + path + ": missing command line")
        }
        
        command := scanner.Text()
        commands = append(commands, command)
    }
    
    return urls, commands, nil
}
