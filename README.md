CppGitMining
============

This go module contains the implementation of a
mining tool for c and cpp git repositories.

Installation:
-------------

An installation of Go (version 1.18 or newer) is necessary,
which can be found at <https://go.dev/>.

The program can be installed using

    > go install ./cmd/cgm
    
Which will save the executable in GOPATH/bin.
The value of GOPATH can be retrieved with

    > go env GOPATH

In order to use the static coupling or change coupling analysis
the StaticCouplingTool (https://github.com/lukas0820/StaticCouplingTool)
or GitCouplingTool (https://github.com/svnlib/GitCouplingTool)
need to be installed and executable from the command line.
Since the GitCouplingTool is run as a jar using java it
is recommended to create a bash-shortcut or to
make an executable script, for instance

    > #!/bin/sh
    > java -jar <path>/GitCouplingTool*.jar "$@"

Usage:
------

After installation the tool can be run
with the following command:

    > cgm [options] <path>
    
To see a list of available options, run

    > cgm -help

The tool requires a file containing a list of
git urls and corresponding build commands as input.
Building a repository is only required for the
static coupling analysis and the compilation database
must be in the root directory of the repository.
For example, if the file "example.txt" contains the lines

    > https://github.com/bitkeeper-scm/bitkeeper.git
    > bear -- make
    
the tool can be run using

    > cgm ./example.txt

Visualisation:
--------------

The jupyter notebook "Render.ipynb" provided in the directory "tools"
can be used to display the results of a program run.
In the first cell of the notebook the variable "resultFile" has to be
set to the path to the result file and then the cell can be run.
Afterwards the other cells can be configured to output different
informations about the results.