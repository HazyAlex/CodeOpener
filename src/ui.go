package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
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

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))
)

type Model struct {
	editorPath string
	projects   []Entry
	shortcuts  []Link

	paginator paginator.Model
	selected  map[int]bool
	cursor    int
	completed bool

	confirmButtonIndex int
}

func (model Model) Init() tea.Cmd {
	return nil
}

func InitModel(editorPath string, projects []Entry, shortcuts []Link) Model {
	const perPage = 10

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = perPage
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(len(projects))

	model := Model{
		editorPath: editorPath,
		projects:   projects,
		shortcuts:  shortcuts,

		paginator: p,
		selected:  make(map[int]bool, len(projects)),

		confirmButtonIndex: len(projects),
	}

	// Initialize the selections based on the existing shortcuts.
	for index, project := range model.projects {
		for _, shortcut := range model.shortcuts {
			if shortcut.Label == project.Label {
				model.selected[index] = true

				break
			}
		}
	}

	return model
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	model.paginator, cmd = model.paginator.Update(msg)

	page := model.paginator.Page // 0-index
	itemsPerPage := model.paginator.PerPage
	itemsOnPage := model.paginator.ItemsOnPage(len(model.projects))

	start := page * itemsPerPage
	end := page*itemsPerPage + itemsPerPage - 1
	if model.paginator.OnLastPage() {
		end = start + itemsOnPage - 1
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "esc", "ctrl+c", "q":
			return model, tea.Quit

		case "up", "k", "shift+tab":
			if model.cursor == model.confirmButtonIndex {
				model.cursor = end
				break
			}
			if model.cursor == start {
				model.cursor = model.confirmButtonIndex
				break
			}
			model.cursor -= 1

		case "down", "j", "tab":
			if model.cursor == model.confirmButtonIndex {
				model.cursor = start
				break
			}
			if model.cursor == end {
				model.cursor = model.confirmButtonIndex
				break
			}
			model.cursor += 1

		case "left", "h", "pgup":
			model.cursor = start

		case "right", "l", "pgdown":
			model.cursor = start

		case "enter", " ":
			if model.cursor != model.confirmButtonIndex {
				model.selected[model.cursor] = !model.selected[model.cursor]
				break
			}

			Submit(model)
			model.completed = true
		}
	}

	return model, cmd
}

func (model Model) View() string {
	var b strings.Builder

	if model.completed {
		b.WriteString(successStyle.Render("Completed successfully!"))
		b.WriteString("\n")
		b.WriteString(successStyle.Render("Press escape (or ctrl+c) to exit!"))
		b.WriteString("\n")

		return b.String()
	}

	b.WriteString(label.Render("Select the projects that should be added as a shortcut."))
	b.WriteString("\n")
	b.WriteString(faintLabel.Render("Entries that are left unchecked will be removed."))
	b.WriteString("\n\n")

	RenderPaginator(&b, model)
	b.WriteString("\n\n")

	if model.cursor == model.confirmButtonIndex {
		b.WriteString(confirmButtonStyle.Render("> Confirm <"))
	} else {
		b.WriteString(confirmButtonStyle.Render("Confirm"))
	}

	return b.String()
}

func RenderPaginator(b *strings.Builder, model Model) {
	start, end := model.paginator.GetSliceBounds(len(model.projects))
	for i, choice := range model.projects[start:end] {
		index := start + i

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

		// Render the checkbox
		label := choice.Label
		folder := displayFolderPath(choice.Folder)

		b.WriteString(checkboxStyle.Render(fmt.Sprintf("%s [", cursor)))
		b.WriteString(checkedStyle.Render(checked))
		b.WriteString(checkboxStyle.Render(fmt.Sprintf("] %s", label)))
		b.WriteString(faintLabel.Render(fmt.Sprintf(" (%s)", folder)))
		b.WriteString("\n")
	}
	b.WriteString("  " + model.paginator.View())
}

func Submit(model Model) error {
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

		err := CreateShortcut(model.editorPath, project.Folder, project.Label)
		if err != nil {
			return err
		}
	}

	return nil
}

func loop(model Model) {
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
