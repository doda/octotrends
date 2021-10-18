package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func loadJSONMap() map[string]RepoInfo {
	data := make(map[string]RepoInfo)

	jsonBytes, err := ioutil.ReadFile(JSONFILENAME)
	if err != nil {
		log.Println("Error loading JSON", err)
	} else {
		json.Unmarshal(jsonBytes, &data)
	}
	return data
}

func writeJSONMap(data map[string]RepoInfo) {
	bytes, _ := json.Marshal(data)
	if err := ioutil.WriteFile(JSONFILENAME, bytes, 0644); err != nil {
		log.Println(err)
	}
}

func GH(repoNames []string) map[string]RepoInfo {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "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"},
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
			jsonMap[repoName] = RepoInfo{}
			continue
		}
		stars := 0
		if repo.StargazersCount != nil {
			stars = *repo.StargazersCount
		}
		language := ""
		if repo.Language != nil {
			language = *repo.Language
		}
		description := ""
		if repo.Description != nil {
			description = *repo.Description
		}
		ri := RepoInfo{
			stars,
			language,
			repo.Topics,
			description,
		}
		jsonMap[repoName] = ri
	}
	writeJSONMap(jsonMap)
	return jsonMap
}
