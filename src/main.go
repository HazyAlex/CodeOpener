package main

import (
	"os"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if runtime.GOOS != "windows" {
		panic("This OS is not supported.")
	}

	if len(os.Args) <= 1 {
		panic("An argument is missing: editor (e.g. 'vscode', 'vscodium').")
	}
	editor := os.Args[1]

	editorExecutablePath, err := GetEditorExecutablePath(editor)
	if err != nil {
		panic(err)
	}

	databasePath, err := GetSQLiteDatabasePath(editor)
	if err != nil {
		panic(err)
	}

	history, err := getRecentlyOpenedProjects(databasePath)
	if err != nil {
		panic(err)
	}

	shortcuts, err := GetExistingShortcuts()
	if err != nil {
		panic(err)
	}

	projects := parseRecentlyOpenedProjects(history)

	// TODO: Pagination
	projects = projects[0:15]

	loop(Model{
		editor:    editorExecutablePath,
		projects:  projects,
		shortcuts: shortcuts,
		selected:  make(map[int]bool, len(projects)),
	})
}
