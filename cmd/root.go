package cmd

import (
	"fmt"
	"github.com/ashishgalagali/go-git-churn/metrics"
	"github.com/spf13/cobra"
	//"io/ioutil"
	"os"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&repoUrl, "repo", "r", "", "Git Repository URL on which the churn metrics has to be computed")
	//print.CheckIfError(cobra.MarkFlagRequired(pf, "repo"))

	//TODO: Enhancements
	//pf.StringVarP(&commitId, "commit", "c", "", "Commit hash for which the metrics has to be computed")
	////print.CheckIfError(cobra.MarkFlagRequired(pf, "commit"))
	//pf.StringVarP(&filepath, "filepath", "f", "", "File path for the file on which the commit metrics has to be computed")
	//pf.StringVarP(&aggregate, "aggregate", "a", "", "Aggregate the churn metrics. \"commit\": Aggregates all files in a commit. \"all\": Aggregate all files all commits and all files")
	//pf.BoolVarP(&whitespace, "whitespace", "w", true, "Excludes whitespaces while calculating the churn metrics is set to false")
	//pf.BoolVarP(&jsonOPToFile, "json", "j", false, "Writes the JSON output to a file within a folder named churn-details")
	//pf.BoolVarP(&printOP, "print", "p", true, "Prints the output in a human readable format")
	//pf.BoolVarP(&enableLog, "logging", "l", false, "Enables logging. Defaults to false")
}

var (
	repoUrl  string
	commitId string
	filepath string
	//whitespace   bool
	//jsonOPToFile bool
	//printOP      bool
	//aggregate    string
	//enableLog    bool

	rootCmd = &cobra.Command{
		Use:   "git-churn",
		Short: "A fast tool for collecting code churn metrics from git repositories.",
		Long: `git-churn gives the churn metrics like insertions, deletions, etc for the given commit hash in the repo specified.
               Complete documentation is available at https://github.com/andymeneely/git-churn`,
		Run: func(cmd *cobra.Command, args []string) {
			//if !enableLog {
			//	helper.INFO.SetFlags(0)
			//	helper.INFO.SetOutput(ioutil.Discard)
			//}
			//helper.INFO.Println("\n Processing new request")
			//helper.INFO.Println("")
			//var churnMetrics interface{}
			var err error

			//commitIds := strings.Split(commitId, "..")
			//firstCommitId := commitIds[0]
			//var secondCommitId = ""
			//if len(commitIds) == 2 {
			//	secondCommitId = strings.TrimFunc(commitIds[1], func(r rune) bool {
			//		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			//	})
			//}
			//if secondCommitId == "" {
			//	commitId = firstCommitId
			//	firstCommitId = ""
			//} else {
			//	commitId = secondCommitId
			//}

			if repoUrl == "" {
				repoUrl = "."
			}
			//repo := metrics.GetRepo(repoUrl)
			//print.PrintInBlue(repoUrl + " " + commitId + " " + filepath + " " + firstCommitId)
			//helper.INFO.Println("Generating git-churn for the following: \n" + "Repo:" + repoUrl + " " + " commitId:" + commitId + " " + " filepath:" + filepath + " " + " firstCommitId:" + firstCommitId)

			//r := metrics.Checkout("https://github.com/ashishgalagali/SWEN610-project", "7368d5fcb7eec950161ed9d13b55caf5961326b6")
			//
			//h, err := r.ResolveRevision(plumbing.Revision("7368d5fcb7eec950161ed9d13b55caf5961326b6"))
			//CheckIfError(err)
			commitObj := metrics.LastCommit(repoUrl)
			CheckIfError(err)
			_, err = metrics.Blame(commitObj, "")

			CheckIfError(err)

			//fmt.Println(fmt.Sprintf("%v", churnMetrics))

		},
	}
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of git-churn",
	Long:  `All software has versions. This is git-churn's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("git-churn version 0.1")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
