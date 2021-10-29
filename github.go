package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func setUpGHClient(githubToken string) (context.Context, *github.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return ctx, github.NewClient(tc)
}

type GithubResult struct {
	name string
	repo *github.Repository
}

// Get additional per-repo information (current # of stars, primary programming language, description)
// directly from GitHub
func RepoWorker(ctx context.Context, client *github.Client, jobs <-chan string, results chan<- GithubResult) {
	for repoName := range jobs {
		log.Println("Getting", repoName)
		nameParts := strings.Split(repoName, "/")
		owner, name := nameParts[0], nameParts[1]
		var repo *github.Repository
		var err error
		// Loop until we're not timed out
		for {
			repo, _, err = client.Repositories.Get(ctx, owner, name)
			if err != nil {
				log.Println(err)
				if _, ok := err.(*github.RateLimitError); ok {
					time.Sleep(time.Minute)
					continue
				}
			}
			break
		}
		if repo == nil {
			log.Println("Repo is nil:", repoName)
			results <- GithubResult{repoName, &github.Repository{}}
		} else {
			results <- GithubResult{repoName, repo}
		}
	}
}

func GetGHRepoInfo(data DataTable, githubToken string, nProc int) map[string]github.Repository {
	ctx, client := setUpGHClient(githubToken)
	GHInfoMap := make(map[string]github.Repository)

	numJobs := len(data)
	jobs := make(chan string, numJobs)
	results := make(chan GithubResult, numJobs)

	for i := 0; i < nProc; i++ {
		go RepoWorker(ctx, client, jobs, results)
	}
	for repo := range data {
		jobs <- repo
	}

	close(jobs)
	for i := 0; i < numJobs; i++ {
		result := <-results
		GHInfoMap[result.name] = *result.repo
	}
	return GHInfoMap
}
