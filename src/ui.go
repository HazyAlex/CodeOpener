package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	label = lipgloss.NewStyle()

	faintLabel = lipgloss.NewStyle().
			Faint(true)

	checkedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	checkboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	confirmButtonStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("69")).
				Padding(0, 2)
)

type Model struct {
	editor    string
	projects  []Entry
	shortcuts []Link

	cursor    int
	selected  map[int]bool
	completed bool
}

func (model Model) Init() tea.Cmd {
	// Initialize the selections based on the existing shortcuts.
	for index, project := range model.projects {
		for _, shortcut := range model.shortcuts {
			if shortcut.Label == project.Label {
				model.selected[index] = true

				break
			}
		}
	}

	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	confirmButtonIndex := len(model.projects) - 1

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "esc", "ctrl+c", "q":
			return model, tea.Quit

		case "up", "k", "shift+tab":
			if model.cursor > 0 {
				model.cursor -= 1
			} else {
				model.cursor = confirmButtonIndex
			}

		case "down", "j", "tab":
			if model.cursor < confirmButtonIndex {
				model.cursor += 1
			} else {
				model.cursor = 0
			}

		case "enter", " ":
			if model.cursor == confirmButtonIndex {
				Apply(model)

				model.completed = true
				break
			}

			model.selected[model.cursor] = !model.selected[model.cursor]
		}
	}

	return model, nil
}

func (model Model) View() string {
	if model.completed {
		return "\nCompleted!\nPress q (or ctrl+c) to exit!\n"
	}

	var b strings.Builder
	b.WriteString(label.Render("Select the projects that should be added as a shortcut."))
	b.WriteString("\n")
	b.WriteString(faintLabel.Render("Entries that are left unchecked will be removed."))
	b.WriteString("\n\n")

	for index, choice := range model.projects {
		// Is the cursor pointing at this choice?
		cursor := " "
		if model.cursor == index {
			cursor = ">"
		}

		// Is this choice selected?
		checked := " "
		if model.selected[index] {
			checked = "x"
		}

		if index < len(model.projects)-1 {
			b.WriteString(checkboxStyle.Render(fmt.Sprintf("%s [", cursor)))
			b.WriteString(checkedStyle.Render(checked))
			b.WriteString(checkboxStyle.Render(fmt.Sprintf("] %s", choice.Label)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if model.cursor == len(model.projects)-1 {
		b.WriteString(confirmButtonStyle.Render("> Confirm <"))
	} else {
		b.WriteString(confirmButtonStyle.Render("Confirm"))
	}

	return b.String()
}

func Apply(model Model) {
	for i, selected := range model.selected {
		project := model.projects[i]

		if !selected {
			// Remove the shortcut if it exists.
			for _, shortcut := range model.shortcuts {
				if shortcut.Label == project.Label {
					RemoveShortcut(project.Label)
					break
				}
			}

			continue
		}

		err := CreateShortcut(model.editor, project.Folder, project.Label)
		if err != nil {
			panic(err)
		}
	}
}

func loop(model Model) {
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
