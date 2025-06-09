package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type Link struct {
	Path  string
	Label string
}

func GetShortcutsPath() (string, error) {
	path := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("Folder: " + path + " does not exist.")
	}

	path = filepath.Join(path, "CodeOpener")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", errors.New("Failed to create the directory " + path + ": " + err.Error())
		}
	}

	return path, nil
}

func GetEditorPath(editor string) (string, error) {
	var path string

	switch strings.ToLower(editor) {
	case "codium", "vscodium":
		path = filepath.Join(os.Getenv("APPDATA"), "VSCodium")
	case "code", "vscode":
		path = filepath.Join(os.Getenv("APPDATA"), "Code")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("Folder: " + path + " does not exist.")
	}

	return path, nil
}

func GetEditorExecutablePath(editor string) (string, error) {
	var path string

	switch strings.ToLower(editor) {
	case "codium", "vscodium":
		path = filepath.Join(os.Getenv("PROGRAMFILES"), "VSCodium", "VSCodium.exe")
	case "code", "vscode":
		path = filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Microsoft VS Code")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("Folder: " + path + " does not exist.")
	}

	return path, nil
}

func GetSQLiteDatabasePath(editor string) (string, error) {
	editorPath, err := GetEditorPath(editor)
	if err != nil {
		return "", err
	}

	path := filepath.Join(editorPath, "User", "globalStorage", "state.vscdb")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("File: " + path + " does not exist.")
	}

	return path, nil
}

func GetExistingShortcuts() ([]Link, error) {
	shortcutsPath, err := GetShortcutsPath()
	if err != nil {
		return nil, err
	}

	var shortcuts []string

	err = filepath.Walk(shortcutsPath, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) != ".lnk" {
			return nil
		}

		shortcuts = append(shortcuts, path)

		return nil
	})

	entries := make([]Link, len(shortcuts))

	for i, entry := range shortcuts {
		entries[i] = Link{
			Path:  entry,
			Label: strings.TrimSuffix(filepath.Base(entry), filepath.Ext(entry)),
		}
	}

	return entries, err
}

func CreateShortcut(target string, link string, name string) error {
	shortcutsPath, err := GetShortcutsPath()
	if err != nil {
		return err
	}

	shortcut := filepath.Join(shortcutsPath, name+".lnk")

	if _, err := os.Stat(shortcut); !os.IsNotExist(err) {
		err = RemoveShortcut(name)
		if err != nil {
			return err
		}
	}

	return CreateWindowsShortcut(
		shortcut,
		target,
		"--folder-uri "+link,
		"Shortcut for project "+name,
	)
}

func CreateWindowsShortcut(shortcut, targetPath, arguments, description string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)

	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()

	shell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer shell.Release()

	cs, err := oleutil.CallMethod(shell, "CreateShortcut", shortcut)
	if err != nil {
		return err
	}

	dispatch := cs.ToIDispatch()
	_, err = oleutil.PutProperty(dispatch, "TargetPath", targetPath)
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(dispatch, "Arguments", arguments)
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(dispatch, "Description", description)
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(dispatch, "Hotkey", "")
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(dispatch, "WindowStyle", "1")
	if err != nil {
		return err
	}
	_, err = oleutil.PutProperty(dispatch, "WorkingDirectory", "")
	if err != nil {
		return err
	}
	_, err = oleutil.CallMethod(dispatch, "Save")
	if err != nil {
		return err
	}

	return nil
}

func RemoveShortcut(name string) error {
	shortcutsPath, err := GetShortcutsPath()
	if err != nil {
		return err
	}

	shortcut := filepath.Join(shortcutsPath, name+".lnk")

	return os.Remove(shortcut)
}
