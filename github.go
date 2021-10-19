package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const JSONFILENAME = "data/repo-info.json"

var GHP = os.Getenv("GHP")

type RepoInfo struct {
	ID           int
	CreatedAt    string
	UpdatedAt    string
	PushedAt     string
	Forks        int
	OpenIssues   int
	NetworkCount int
	Subscribers  int
	FullName     string
	Stars        int
	Language     string
	Topics       []string
	Description  string
}

func loadJSONMap() map[string]github.Repository {
	data := make(map[string]github.Repository)

	jsonBytes, err := ioutil.ReadFile(JSONFILENAME)
	if err != nil {
		log.Println("Error loading JSON", err)
	} else {
		json.Unmarshal(jsonBytes, &data)
	}
	return data
}

func writeJSONMap(data map[string]github.Repository) {
	bytes, _ := json.Marshal(data)
	if err := ioutil.WriteFile(JSONFILENAME, bytes, 0644); err != nil {
		log.Println(err)
	}
}

func GetGHRepoInfo(repoNames []string) map[string]github.Repository {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GHP},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	jsonMap := loadJSONMap()
	log.Println("Loaded JSON Map", len(jsonMap))
	for _, repoName := range repoNames {
		if _, ok := jsonMap[repoName]; ok {
			// We have this info skip it
			log.Printf("Skipping %s", repoName)
			continue
		}
		log.Println("Getting", repoName)
		spluts := strings.Split(repoName, "/")
		owner, name := spluts[0], spluts[1]
		var repo *github.Repository
		var err error
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
			jsonMap[repoName] = github.Repository{}
			continue
		}

		jsonMap[repoName] = *repo
	}
	writeJSONMap(jsonMap)
	return jsonMap
}
