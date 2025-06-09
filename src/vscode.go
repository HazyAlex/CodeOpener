package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"path/filepath"
	"strings"
)

type Entry struct {
	Folder string `json:"folderUri"`
	Label  string `json:"label"`
}

func getRecentlyOpenedProjects(sqlFile string) (string, error) {
	db, err := sql.Open("sqlite", sqlFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT * FROM ItemTable`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			key   string
			value string
		)

		if err := rows.Scan(&key, &value); err != nil {
			panic(err)
		}

		if key == "history.recentlyOpenedPathsList" {
			return value, nil
		}
	}

	return "", errors.New("could not find any recently opened projects")
}

func parseRecentlyOpenedProjects(history string) []Entry {
	var result map[string][]Entry
	if err := json.Unmarshal([]byte(history), &result); err != nil {
		panic(err)
	}

	var folders []Entry

	for _, entry := range result["entries"] {
		if entry.Folder != "" {
			folders = append(folders, Entry{
				Folder: entry.Folder,
				Label:  filepath.Base(entry.Folder),
			})
		}
	}

	return folders
}

func displayFolderPath(path string) string {
	symbols := []string{"vscode-remote://", "file://"}

	path, err := url.QueryUnescape(path)
	if err != nil {
		panic(err)
	}

	for _, symbol := range symbols {
		path = strings.TrimLeft(path, symbol)
	}

	return path
}
