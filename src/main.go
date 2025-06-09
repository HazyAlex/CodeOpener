package main

import (
	"os"
	"runtime"

	_ "modernc.org/sqlite"
)

func main() {
	if runtime.GOOS != "windows" {
		panic("This OS is not supported.")
	}

	if len(os.Args) <= 1 {
		panic("An argument is missing: editor (e.g. 'vscode', 'vscodium', 'cursor').")
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

	loop(InitModel(editorExecutablePath, projects, shortcuts))
}
