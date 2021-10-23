package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Get additional per-repo information (current # of stars, primary programming language, description)
// directly from GitHub
func GetGHRepoInfo(data DataTable, GitHubToken string) map[string]github.Repository {
	// Set up GH API client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	GHInfoMap := make(map[string]github.Repository)

	for repoName := range data {
		log.Println("Getting", repoName)
		nameParts := strings.Split(repoName, "/")
		owner, name := nameParts[0], nameParts[1]
		var repo *github.Repository
		var err error
		// Loop until we're not timed out
		for {
			repo, _, err = client.Repositories.Get(ctx, owner, name)
			if _, ok := err.(*github.RateLimitError); ok {
				log.Println("Hit rate limit, sleeping 1 minute")
				time.Sleep(time.Minute)
			} else {
				break
			}
		}
		if repo == nil {
			log.Println("Repo is nil:", repoName)
			GHInfoMap[repoName] = github.Repository{}
			continue
		}

		GHInfoMap[repoName] = *repo
	}
	return GHInfoMap
}
