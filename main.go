package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/google/go-github/github"

	"github.com/jmoiron/sqlx"
)

type DataTable map[string]TableItem

type TableItem struct {
	Added10 int
	Added30 int
	Added90 int
}

type JSONOutItem struct {
	Name        string
	Stars       int
	Added10     int
	Added30     int
	Added90     int
	Language    string
	Topics      string
	Description string
}

var StarsSelectQuery = `
WITH 
	(SELECT max(created_at) FROM github_events) as right_now,
	dateDiff('day', created_at_sub, right_now) as days
SELECT
	repo_name,
	sum(days <= 10) as added10,
	sum(days <= 30) as added30,
	sum(days <= 90) as added90
FROM 
( // at most 1 star per user per repo
	SELECT
		repo_name,
		actor_login,
		any(created_at) as created_at_sub
	FROM github_events
	WHERE event_type = 'WatchEvent'
	GROUP BY repo_name, actor_login
)
WHERE repo_name in 
( // "interesting" repos list
	SELECT
		repo_name
	FROM github_events
	WHERE event_type = 'WatchEvent' AND created_at > minus(right_now, toIntervalDay(?))
	GROUP BY repo_name
	ORDER BY count() DESC
	LIMIT ?
)
GROUP BY repo_name
`

// Use ClickHouse to quickly aggregate how many people have starred the repo recently
func GetGrowths(connect *sqlx.DB, lookback int, numRepos int) (DataTable, error) {
	data := DataTable{}

	var items []struct {
		RepoName string `db:"repo_name"`
		Added10  int32  `db:"added10"`
		Added30  int32  `db:"added30"`
		Added90  int32  `db:"added90"`
	}

	log.Println("Running", StarsSelectQuery)
	if err := connect.Select(&items, StarsSelectQuery, lookback, numRepos); err != nil {
		return nil, err
	}

	for _, item := range items {
		dataItem := data[item.RepoName]
		dataItem.Added10 = int(item.Added10)
		dataItem.Added30 = int(item.Added30)
		dataItem.Added90 = int(item.Added90)
		data[item.RepoName] = dataItem
	}
	log.Printf("Aggregated %d items", len(items))
	return data, nil
}

// Write our accumulated & joined information out to JSON for the frontend to consume
func WriteToJSON(d DataTable, jsonMap map[string]github.Repository, outFileName string) error {
	outItems := []JSONOutItem{}
	for repoName, tableItem := range d {
		repoInfo := jsonMap[repoName]

		language := StringValue(repoInfo.Language)
		if RepoLangDoesntCount(repoName) {
			language = ""
		}
		outItems = append(outItems, JSONOutItem{
			Name:        repoName,
			Stars:       IntValue(repoInfo.StargazersCount),
			Added10:     tableItem.Added10,
			Added30:     tableItem.Added30,
			Added90:     tableItem.Added90,
			Language:    language,
			Topics:      strings.Join(repoInfo.Topics, ", "),
			Description: StringValue(repoInfo.Description),
		})
	}
	bytes, err := json.Marshal(outItems)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outFileName, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	numRepos := flag.Int("numrepos", 5000, "Number of repos to get")
	lookback := flag.Int("lookback", 30, "Number of days to lookback to determine which repos to get")
	clickHouseURL := flag.String("clickhouse", "tcp://play.clickhouse.com:9440?debug=false&username=play&secure=true", "ClickHouse TCP URL")
	githubToken := flag.String("ghp", "", "GitHub Access token")
	nProc := flag.Int("n", 1, "Number of worker processes")
	outFileName := flag.String("o", "out.json", "Output file name")

	flag.Parse()

	if *githubToken == "" {
		log.Println("Requires a GitHub Access Token")
		return
	}

	log.Printf("Getting %d repos that have received the most stars in the past %d days, and using ClickHouse URL: %s\n", *numRepos, *lookback, *clickHouseURL)

	connect, err := sqlx.Open("clickhouse", *clickHouseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Get Growth dataTable from ClickHouse
	dataTable, err := GetGrowths(connect, *lookback, *numRepos)
	if err != nil {
		log.Fatal(err)
	}

	// Get GitHub dataTable for these repos (either cached or anew)
	GHInfoMap := GetGHRepoInfo(dataTable, *githubToken, *nProc)

	// Write out
	err = WriteToJSON(dataTable, GHInfoMap, *outFileName)
	if err != nil {
		log.Fatal(err)
	}
}
