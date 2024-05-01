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

	folders := parseRecentlyOpenedProjects(history)

	for _, entry := range folders {
		for _, shortcut := range shortcuts {
			if shortcut.Label == entry.Label {
				// TODO: Query if the user wants to update or ignore.
				break
			}
		}

		err := CreateShortcut(editorExecutablePath, entry.Folder, entry.Label)
		if err != nil {
			panic(err)
		}
	}
}
