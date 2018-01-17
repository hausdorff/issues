package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

//
// Polling.
//

func pollGitHub() (*IssueIndex, chan struct{}) {
	// ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})

	// Initial data.
	issues := &IssueIndex{}
	err := issues.Update()
	if err != nil {
		log.Printf("Failed to get issues:\n%v\n", err)
	}

	// Poll.
	// go func() {
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			{
	// 				log.Println("Polling GitHub")
	// 				err := issues.Update()
	// 				if err != nil {
	// 					log.Printf("Failed to get issues:\n%v\n", err)
	// 				}
	// 			}
	// 		case <-quit:
	// 			ticker.Stop()
	// 			return
	// 		}
	// 	}
	// }()

	return issues, quit
}

//
// ksonnet org config data.
//

const (
	ksonnetOrg = "ksonnet"

	ksonnetRepo    = "ksonnet"
	ksonnetLibRepo = "ksonnet-lib"
	partsRepo      = "parts"
	clientRepo     = "client"

	untriaged = ""
	bugs      = "bug"
)

var repos = []string{ksonnetRepo, ksonnetLibRepo, partsRepo, clientRepo}

//
// GitHub client managers.
//

func makeGithubClient() *github.Client {
	var hc *http.Client

	ght := os.Getenv("GITHUB_TOKEN")
	if len(ght) > 0 {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ght},
		)
		hc = oauth2.NewClient(ctx, ts)
	}

	return github.NewClient(hc)
}

func ksonnetIssues() (Issues, map[string]Issues, error) {
	client := makeGithubClient()

	ctx := context.Background()
	opts := github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 10000,
		},
	}
	issues, _, err := client.Issues.ListByRepo(ctx, ksonnetOrg, ksonnetRepo, &opts)
	if err != nil {
		return nil, nil, err
	}

	return issues, labelNameHistogram(issues), nil
}

//
// Histogram implementation.
//

type Issues []*github.Issue
type Histogram map[string]Issues
type Snapshot struct {
	data   Histogram
	issues Issues
}
type IssueIndex struct {
	sync.RWMutex
	Snapshot
}

func (idx *IssueIndex) Update() error {
	issues, data, err := ksonnetIssues()
	if err != nil {
		return err
	}
	idx.Lock()
	idx.issues = issues
	idx.data = data
	defer idx.Unlock()
	return nil
}

func (idx *IssueIndex) GetSnapshot() Snapshot {
	idx.RLock()
	defer idx.RUnlock()

	snapshot := Snapshot{data: Histogram{}, issues: Issues{}}
	for label, issues := range idx.data {
		snapshot.data[label] = issues
	}

	for _, issue := range idx.issues {
		snapshot.issues = append(snapshot.issues, issue)
	}

	return snapshot
}

func (idx *IssueIndex) marshal() ([]byte, error) {
	idx.RLock()
	defer idx.RUnlock()
	return json.Marshal(idx.data)
}

type BucketType int

const (
	Hour BucketType = iota
	Day
	Month
	Year
)

func (snap *Snapshot) Bugs() Issues {
	if bugs, exists := snap.data[bugs]; exists {
		return bugs
	}
	return Issues{}
}

func (snap *Snapshot) Untriaged() Issues {
	if bugs, exists := snap.data[untriaged]; exists {
		return bugs
	}
	return Issues{}
}

func (is Issues) CumulativeCount() map[string]int {
	// Lexically-sortable datetime format.
	const dateFmt = "2006-01-02"

	counts := map[string]int{}
	for _, issue := range is {
		created := issue.GetCreatedAt().Format(dateFmt)
		counts[created]++

		if issue.GetState() == "closed" {
			closed := issue.GetClosedAt().Format(dateFmt)
			counts[closed]--
		}
	}

	times := []string{}
	for time := range counts {
		times = append(times, time)
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	curr := 0
	for _, time := range times {
		curr += counts[time]
		counts[time] = curr
	}

	return counts
}

func labelNameHistogram(issues []*github.Issue) map[string]Issues {
	// NOTE: this is not technically correct, since:
	//   1. labels can have the same name. We don't really care about this corner
	//      case, because we will never make two different labels with the same
	//      name.
	//   2. labels can technically have an empty name. We don't really care about
	//      this corner case, either, since we consider labels with no name to be
	//      "untriaged".
	hist := map[string]Issues{}
	for _, issue := range issues {
		// fmt.Println(i, issue.GetState())
		if len(issue.Labels) == 0 {
			hist[untriaged] = append(hist[untriaged], issue)
		}

		for _, label := range issue.Labels {
			name := label.GetName()
			// fmt.Println("\t" + name)
			hist[name] = append(hist[name], issue)
		}
	}

	return hist
}
