# go-git-churn

A fast tool for collecting code churn metrics from git repositories.

# Installation

You will need Go language installed on your system. Ref: https://golang.org/doc/install

```
   git clone https://github.com/ashishgalagali/go-git-churn
   cd go-git-churn
   go install github.com/ashishgalagali/go-git-churn
   go build
 ```

The `--repo` flag takes either github URL of the repo in which case it will clone the repo into the local memory and
performs the operations, or you can specify the path to the cloned repo on your system. Use `"."` if the working
directory is the repo to be used

```
   ./go-git-churn --help
   ./go-git-churn --repo https://github.com/ashishgalagali/SWEN610-project 
   /path/to/go-git-churn --repo /path/to/repo 
```

The output will be written to output_timeStamp.json file in the outputs folder

## Options

```
Flags:
  -h, --help                help for git-churn
  -r, --repo string         Git Repository URL on which the churn metrics has to be computed
  -c, --commit              The Last commit hash till which the metrics has to be computed
  -f, --filepath            File path to filter file on which the churn metrics has to be computed
```

## Future work

1. Track the deleted files