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

func LoadJSONMap() map[string]github.Repository {
	data := make(map[string]github.Repository)

	jsonBytes, err := ioutil.ReadFile(JSONFILENAME)
	if err != nil {
		log.Println("Error loading JSON", err)
	} else {
		json.Unmarshal(jsonBytes, &data)
	}
	return data
}

func WriteJSONMap(data map[string]github.Repository) {
	bytes, _ := json.Marshal(data)
	if err := ioutil.WriteFile(JSONFILENAME, bytes, 0644); err != nil {
		log.Println(err)
	}
}

func GetGHRepoInfo(data DataTable) map[string]github.Repository {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GHP},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	jsonMap := LoadJSONMap()
	log.Println("Loaded JSON Map", len(jsonMap))
	for repoName := range data {
		if _, ok := jsonMap[repoName]; ok {
			// We have this info, skip it
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
	WriteJSONMap(jsonMap)
	return jsonMap
}
