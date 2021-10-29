package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-github/github"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestGetGrowths(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	numRepos := 2
	lookback := 30
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	rows := sqlmock.NewRows([]string{
		"repo_name",
		"added7",
		"added30",
		"added90",
	}).AddRow("test/repo", 10, 20, 30)

	mock.ExpectQuery("SELECT repo_name, (.+) FROM").WithArgs(lookback, numRepos).WillReturnRows(rows)

	data, err := GetGrowths(sqlxDB, lookback, numRepos)
	require.NoError(t, err)

	require.Equal(t, data, DataTable{
		"test/repo": TableItem{10, 20, 30},
	})
	mockDB.Close()
}

func TestWriteJSON(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	data := DataTable{
		"test/repo": TableItem{10, 20, 30},
	}
	language := "Go"
	count := 55
	jsonMap := map[string]github.Repository{
		"test/repo": {
			Language:        &language,
			StargazersCount: &count,
			Topics:          []string{"a", "b"},
		},
	}

	fName := dir + "/test.json"
	err = WriteToJSON(data, jsonMap, fName)
	require.NoError(t, err)

	dat, _ := os.ReadFile(fName)
	require.Equal(t, string(dat),
		`[{"Name":"test/repo","Stars":55,"Added7":10,"Added30":20,"Added90":30,"Language":"Go","Topics":"a, b","Description":""}]`,
	)
}
