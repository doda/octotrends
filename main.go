package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
)

const MINSTARSLASTYEAR = 1000

type Period struct {
	name string
	days int
}

type DataTable map[string]TableItem

func (d DataTable) Keys() []string {
	keys := make([]string, 0, len(d))

	for k := range d {
		keys = append(keys, k)
	}
	return keys
}

type Growther struct {
	StarsPenult int
	StarsUlt    int
	Growth      float64
}

type TableItem struct {
	StarsGrowth map[int]Growther
}

func getRepos(connect *sqlx.DB) DataTable {
	var items []string

	selector := `
	SELECT
		repo_name
	FROM github_events
	WHERE event_type = 'WatchEvent' AND created_at > minus(now(), toIntervalDay(365))
	GROUP BY repo_name
	HAVING (count() >= ?);
	`

	if err := connect.Select(&items, selector, MINSTARSLASTYEAR); err != nil {
		log.Fatal(err)
	}

	data := DataTable{}
	for _, item := range items {
		log.Printf("name: %s", item)
		data[item] = TableItem{map[int]Growther{}}
	}
	log.Printf("# items: %d", len(items))
	return data
}

func getGrowths(connect *sqlx.DB, data DataTable) {
	periods := []Period{
		{"1y", 365},
		{"6mo", 180},
		{"1mo", 30},
	}
	selector := `
	WITH dateDiff('day', created_at, (SELECT max(created_at) FROM github_events)) as days
	SELECT
		repo_name,
		sum(days <= ? * 2 and days > ?) as penult,
		sum(days <= ?) as ult,
		round(ult / penult, 3) as growth
	FROM github_events
	WHERE event_type = 'WatchEvent'
	GROUP BY repo_name
	HAVING (max(days) >= ? * 2) and penult > 0 and repo_name in (?)
	ORDER BY growth desc
	`
	for _, period := range periods {
		var items []struct {
			RepoName string  `db:"repo_name"`
			Penult   int32   `db:"penult"`
			Ult      int32   `db:"ult"`
			Growth   float64 `db:"growth"`
		}

		if err := connect.Select(&items, selector, period.days, period.days, period.days, period.days, data.Keys()); err != nil {
			log.Fatal(err)
		}

		for _, item := range items {
			log.Printf("name: %s, growth: %f", item.RepoName, item.Growth)
			itemHere := data[item.RepoName]
			itemHere.StarsGrowth[period.days] = Growther{int(item.Penult), int(item.Ult), item.Growth}
		}
		log.Printf("# items: %d", len(items))

	}
}
func WriteToCSV(d DataTable) {
	w := csv.NewWriter(os.Stdout)

	for repoName, tableItem := range d {
		record := []string{repoName, fmt.Sprint(tableItem.StarsGrowth[30].Growth), fmt.Sprint(tableItem.StarsGrowth[180].Growth), fmt.Sprint(tableItem.StarsGrowth[365].Growth)}

		if err := w.Write(record); err != nil {
			log.Fatalln("OMFG", err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	connect, err := sqlx.Open("clickhouse", "tcp://gh-api.clickhouse.tech:9440?debug=true&username=explorer&secure=true")
	if err != nil {
		log.Fatal(err)
	}

	data := getRepos(connect)
	getGrowths(connect, data)
	// fmt.Println(data)
	WriteToCSV(data)
}
