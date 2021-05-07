package metrics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ashishgalagali/go-git-churn/helper"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/diff"
	"strings"
	"time"
)

// BlameResult represents the result of a Blame operation.
type BlameResult struct {
	// Path is the path of the File that we're blaming.
	Path string
	// Rev (Revision) is the hash of the specified Commit used to generate this result.
	Rev plumbing.Hash
	// Lines contains every line with its authorship.
	Lines  []*Line
	Churns []Churn
}

// Blame returns a BlameResult with the information about the last author of
// each line from file `path` at commit `c`.
func Blame(c *object.Commit, path string, lastCommitId string) (*BlameResult, error) {
	// The file to blame is identified by the input arguments:
	// commit and path. commit is a Commit object obtained from a Repository. Path
	// represents a path to a specific file contained into the repository.
	//
	// Blaming a file is a two step process:
	//
	// 1. Create a linear history of the commits affecting a file. We use
	// revlist.New for that.
	//
	// 2. Then build a graph with a node for every line in every file in
	// the history of the file.
	//
	// Each node is assigned a commit: Start by the nodes in the first
	// commit. Assign that commit as the creator of all its lines.
	//
	// Then jump to the nodes in the next commit, and calculate the diff
	// between the two files. Newly created lines get
	// assigned the new commit as its origin. Modified lines also get
	// this new commit. Untouched lines retain the old commit.
	//
	// All this work is done in the assignOrigin function which holds all
	// the internal relevant data in a "blame" struct, that is not
	// exported.
	//
	// TODO: ways to improve the efficiency of this function:
	// 1. Improve revlist
	// 2. Improve how to traverse the history (example a backward traversal will
	// be much more efficient)
	//
	// TODO: ways to improve the function in general:
	// 1. Add memoization between revlist and assign.
	// 2. It is using much more memory than needed, see the TODOs below.

	b := new(blame)
	b.fRev = c
	//b.pRev = p
	// TODO: filter is path is not empty
	b.path = path
	b.lastCommitId = lastCommitId

	// get all the file revisions
	if err := b.fillRevs(); err != nil {
		return nil, err
	}

	// calculate the line tracking graph and fill in
	// file contents in data.
	if err := b.fillGraphAndData(); err != nil {
		return nil, err
	}

	//file, err := b.fRev.File(b.path)
	//if err != nil {
	//	return nil, err
	//}
	//finalLines, err := file.Lines()
	//if err != nil {
	//	return nil, err
	//}

	// Each node (line) holds the commit where it was introduced or
	// last modified. To achieve that we use the FORWARD algorithm
	// described in Zimmermann, et al. "Mining Version Archives for
	// Co-changed Lines", in proceedings of the Mining Software
	// Repositories workshop, Shanghai, May 22-23, 2006.
	//lines, err := newLines(finalLines, b.sliceGraph(len(b.graph)-1))
	//if err != nil {
	//	return nil, err
	//}

	Churns := make([]Churn, len(b.revs))
	for i := 0; i < len(b.revs); i++ {
		Churns[i] = Churn{
			CommitID:      b.revs[i].Hash.String(),
			CommitAuthor:  b.revs[i].Author.Email,
			Date:          b.revs[i].Author.When.String(),
			CommitMessage: b.revs[i].Message,
			ChurnFiles:    b.ChurnFiles[i],
		}
	}

	return &BlameResult{
		Path: path,
		Rev:  c.Hash,
		//Lines:  lines,
		Churns: Churns,
	}, nil
}

type ChurnFile struct {
	FileName  string
	SelfChurn []int
	//TODO:
	InteractiveChurn map[string][]int // Hash of authors and count
}

type Churn struct {
	CommitID      string
	CommitAuthor  string
	Date          string
	CommitMessage string
	ChurnFiles    []ChurnFile
}

// Line values represent the contents and author of a line in BlamedResult values.
type Line struct {
	// Author is the email address of the last author that modified the line.
	Author string
	// Text is the original text of the line.
	Text string
	// Date is when the original text of the line was introduced
	Date time.Time
	// Hash is the commit hash that introduced the original line
	Hash plumbing.Hash
}

func newLine(author, text string, date time.Time, hash plumbing.Hash) *Line {
	return &Line{
		Author: author,
		Text:   text,
		Hash:   hash,
		Date:   date,
	}
}

func newLines(contents []string, commits []*object.Commit) ([]*Line, error) {
	lcontents := len(contents)
	lcommits := len(commits)

	if lcontents != lcommits {
		if lcontents == lcommits-1 && contents[lcontents-1] != "\n" {
			contents = append(contents, "\n")
		} else {
			return nil, errors.New("contents and commits have different length")
		}
	}

	result := make([]*Line, 0, lcontents)
	for i := range contents {
		result = append(result, newLine(
			commits[i].Author.Email, contents[i],
			commits[i].Author.When, commits[i].Hash,
		))
	}

	return result, nil
}

// this struct is internally used by the blame function to hold its
// inputs, outputs and state.
type blame struct {
	// the path of the file to blame
	path         string
	lastCommitId string
	// the commit of the final revision of the file to blame
	fRev *object.Commit

	// the commit of the parent revision of the file to blame till
	//pRev *object.Commit

	// the chain of revisions affecting the the file to blame
	revs []*object.Commit
	// the contents of the file across all its revisions
	//data []string
	data map[string][]string

	// the graph of the lines in the file across all the revisions
	graph map[string][][]*object.Commit

	commitIndexMap map[string]int

	ChurnFiles [][]ChurnFile
}

// calculate the history of a file "path", starting from commit "from", sorted by commit date.
func (b *blame) fillRevs() error {
	var err error

	b.revs, err = references(b.fRev, b.path)
	return err
}

// build graph of a file from its revision history
func (b *blame) fillGraphAndData() error {
	//TODO: not all commits are needed, only the current rev and the prev
	//b.graph = make([][]*object.Commit, len(b.revs))
	b.graph = make(map[string][][]*object.Commit)
	//b.data = make([]string, len(b.revs)) // file contents in all the revisions
	b.data = make(map[string][]string)
	b.ChurnFiles = make([][]ChurnFile, len(b.revs))
	b.commitIndexMap = make(map[string]int)

	for i, rev := range b.revs {
		b.commitIndexMap[rev.Hash.String()] = i
	}

	var opFileName = "outputs/output_" + time.Now().UTC().Format("2006-01-02T15:04:05-0700") + ".json"
	// for every revision of the file, starting with the first
	// one...
	helper.AppendToFile(opFileName, "[")
	for i, rev := range b.revs {
		//cTree, _ := rev.Tree()
		//if rev.Hash.String() == "e15b720263903680264fdfb124749b6f386d51e6" {
		//	fmt.Println("STOP")
		//}
		//	pTree, _ := b.revs[i-1].Tree()
		//	changes, _ := cTree.Diff(pTree)
		//	print(changes)
		//}
		if b.lastCommitId != "" && b.lastCommitId == rev.Hash.String() {
			break
		}

		ittr, _ := rev.Files()
		commitFiles := make([]ChurnFile, 0)
		for {
			file, err := ittr.Next()

			if file == nil {
				break
			}
			if (b.path != "" && b.path == file.Name) || (b.path == "") {
				churnDetails := new(ChurnFile)
				churnDetails.FileName = file.Name
				// get the contents of the file
				//file, err := rev.Filele(b.path)
				if err != nil {
					return nil
				}
				if _, ok := b.data[file.Name]; !ok {
					//do something here
					b.data[file.Name] = make([]string, len(b.revs))
				}
				b.data[file.Name][i], err = file.Contents()
				if err != nil {
					return err
				}
				nLines := countLines(b.data[file.Name][i])
				// create a node for each line
				if _, ok := b.graph[file.Name]; !ok {
					//do something here
					b.graph[file.Name] = make([][]*object.Commit, len(b.revs))
				}
				b.graph[file.Name][i] = make([]*object.Commit, nLines)
				// assign a commit to each node
				// if this is the first revision, then the node is assigned to
				// this first commit.
				if i == 0 {
					for j := 0; j < nLines; j++ {
						b.graph[file.Name][i][j] = b.revs[i]
					}
				} else {
					// if this is not the first commit, then assign to the old
					// commit or to the new one, depending on what the diff
					// says.

					//if strings.Contains(rev.Message, "Merge pull request"){
					//	continue
					//}
					//Setting it to MAX=1
					nearestParent := len(b.revs) + 1
					iter := rev.Parents()
					count := 0
					for {
						parent, _ := iter.Next()
						if parent == nil {
							break
						}
						count++
						//if count > 1 && strings.Contains(parent.Message, "Merge pull request") {
						//	break
						//}
						//if count > 1 {
						//	break
						//}
						parentIndex := b.commitIndexMap[parent.Hash.String()]
						if nearestParent > parentIndex {
							if count > 1 {
								if !strings.Contains(parent.Message, "Merge pull request") {
									nearestParent = parentIndex
								}
							} else {
								nearestParent = parentIndex
							}
						} else if !strings.Contains(parent.Message, "Merge pull request") {
							nearestParent = parentIndex
						}
						//if nearestParent > parentIndex {
						//	nearestParent = parentIndex
						//}
					}
					if count == 1 {
						b.assignOrigin(i, nearestParent, churnDetails, false)
					} else {
						b.assignOrigin(i, nearestParent, churnDetails, true)
					}
				}
				if len(churnDetails.InteractiveChurn) != 0 || len(churnDetails.SelfChurn) != 0 {
					commitFiles = append(commitFiles, *churnDetails)
				}
			}
		}
		//if len(commitFiles) != 0 {
		//b.ChurnFiles[i] = commitFiles
		churn := Churn{
			CommitID:      b.revs[i].Hash.String(),
			CommitAuthor:  b.revs[i].Author.Email,
			Date:          b.revs[i].Author.When.String(),
			CommitMessage: b.revs[i].Message,
			ChurnFiles:    commitFiles,
		}
		data, _ := json.Marshal(churn)
		if i != 0 {
			helper.AppendToFile(opFileName, ",")
		}
		helper.AppendToFile(opFileName, string(data)+"\n")
		//fmt.Printf("%s\n", data)
		//fmt.Println("\n")
		//}
	}
	helper.AppendToFile(opFileName, "]")
	return nil
}

// sliceGraph returns a slice of commits (one per line) for a particular
// revision of a file (0=first revision).
//func (b *blame) sliceGraph(i int) []*object.Commit {
//	fVs := b.graph[i]
//	result := make([]*object.Commit, 0, len(fVs))
//	for _, v := range fVs {
//		c := *v
//		result = append(result, &c)
//	}
//	return result
//}

// Assigns origin to vertexes in current (c) rev from data in its previous (p)
// revision
func (b *blame) assignOrigin(c, p int, churnDetails *ChurnFile, copyAsIs bool) {
	// assign origin based on diff info
	hunks := diff.Do(b.data[churnDetails.FileName][p], b.data[churnDetails.FileName][c])

	sl := -1 // source line
	dl := -1 // destination line
	for h := range hunks {
		hLines := countLines(hunks[h].Text)
		for hl := 0; hl < hLines; hl++ {
			switch {
			case hunks[h].Type == 0:
				sl++
				dl++
				b.graph[churnDetails.FileName][c][dl] = b.graph[churnDetails.FileName][p][sl]
			case hunks[h].Type == 1:
				dl++
				if copyAsIs {
					//if strings.Contains(b.revs[p].Message, "Merge pull request") {
					//	fmt.Println(b.revs[c].Hash.String())
					//}
					b.graph[churnDetails.FileName][c][dl] = b.revs[p]
				} else {
					//if strings.Contains(b.revs[c].Message, "Merge pull request") {
					//	fmt.Println(b.revs[c].Hash.String())
					//}
					b.graph[churnDetails.FileName][c][dl] = b.revs[c]
				}
			case hunks[h].Type == -1:
				sl++
				if b.revs[c].Author.Email == b.graph[churnDetails.FileName][p][sl].Author.Email {
					churnDetails.SelfChurn = append(churnDetails.SelfChurn, sl+1)
				} else {
					ichurn := churnDetails.InteractiveChurn[b.graph[churnDetails.FileName][p][sl].Author.Email]
					if len(ichurn) == 0 {
						churnDetails.InteractiveChurn = make(map[string][]int)
					}
					ichurn = append(ichurn, sl+1)
					churnDetails.InteractiveChurn[b.graph[churnDetails.FileName][p][sl].Author.Email] = ichurn
				}
			default:
				panic("unreachable")
			}
		}
	}
	if c-20 > 0 {
		b.graph[churnDetails.FileName][c-20] = make([]*object.Commit, 0)
		b.data[churnDetails.FileName][c-20] = ""
	}
}

// GoString prints the results of a Blame using git-blame's style.
func (b *blame) GoString() string {
	var buf bytes.Buffer

	//file, err := b.fRev.File(b.path)
	//if err != nil {
	//	panic("PrettyPrint: internal error in repo.Data")
	//}
	//contents, err := file.Contents()
	//if err != nil {
	//	panic("PrettyPrint: internal error in repo.Data")
	//}
	//
	//lines := strings.Split(contents, "\n")
	//// max line number length
	//mlnl := len(strconv.Itoa(len(lines)))
	//// max author length
	//mal := b.maxAuthorLength()
	//format := fmt.Sprintf("%%s (%%-%ds %%%dd) %%s\n",
	//	mal, mlnl)

	//fVs := b.graph[len(b.graph)-1]
	//for ln, v := range fVs {
	//	fmt.Fprintf(&buf, format, v.Hash.String()[:8],
	//		prettyPrintAuthor(fVs[ln]), ln+1, lines[ln])
	//}
	return buf.String()
}

// utility function to pretty print the author.
func prettyPrintAuthor(c *object.Commit) string {
	return fmt.Sprintf("%s %s", c.Author.Name, c.Author.When.Format("2006-01-02"))
}

// utility function to calculate the number of runes needed
// to print the longest author name in the blame of a file.
//func (b *blame) maxAuthorLength() int {
//	memo := make(map[plumbing.Hash]struct{}, len(b.graph)-1)
//	fVs := b.graph[len(b.graph)-1]
//	m := 0
//	for ln := range fVs {
//		if _, ok := memo[fVs[ln].Hash]; ok {
//			continue
//		}
//		memo[fVs[ln].Hash] = struct{}{}
//		m = max(m, utf8.RuneCountInString(prettyPrintAuthor(fVs[ln])))
//	}
//	return m
//}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
