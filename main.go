package main

import (
	"github.com/ashishgalagali/go-git-churn/cmd"
	"runtime"
)

//var cpuprofile = flag.String("cpuprofile", "defaultProf.out", "write cpu profile to `file`")
//var memprofile = flag.String("memprofile", "defaultMem.out", "write memory profile to `file`")

func main() {
	//flag.Parse()
	//if *cpuprofile != "" {
	//	f, err := os.Create(*cpuprofile)
	//	if err != nil {
	//		log.Fatal("could not create CPU profile: ", err)
	//	}
	//	defer f.Close()
	//	if err := pprof.StartCPUProfile(f); err != nil {
	//		log.Fatal("could not start CPU profile: ", err)
	//	}
	//	defer pprof.StopCPUProfile()
	//}
	//For executing the concurrent go routines in the program parallelly
	numcpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numcpu)
	cmd.Execute()

	// Use this to run on the IDE
	//r := metrics.Checkout("https://github.com/ashishgalagali/SWEN610-project", "7368d5fcb7eec950161ed9d13b55caf5961326b6")
	//
	//h, err := r.ResolveRevision(plumbing.Revision("7368d5fcb7eec950161ed9d13b55caf5961326b6"))
	//cmd.CheckIfError(err)
	//commitObj, err := r.CommitObject(*h)
	//cmd.CheckIfError(err)
	//_, err = metrics.Blame(commitObj, "", "3d5168fbca9299add91a28464d9c7586aa66d58f")

}
