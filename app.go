package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guillermo/linear/linear-api"
)

type App struct {
	help     help.Model
	loading  bool
	keys     keyMap
	quitting bool

	spinner spinner.Model

	someData      bool
	issues        Issues
	selectedIssue *linear.Issue

	err error
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding

	Quit key.Binding
	Help key.Binding
}

var keys = keyMap{
	Up:     key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "move up")),
	Down:   key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "move down")),
	Quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	Select: key.NewBinding(key.WithKeys("enter", "x"), key.WithHelp("enter", "select")),
	Help:   key.NewBinding(key.WithKeys("h", "f1"), key.WithHelp("h", "help")),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys.Help, keys.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select},
		k.ShortHelp(),
	}
}

func (a *App) Init() tea.Cmd {
	a.help = help.New()
	a.keys = keys
	a.loading = true
	a.spinner = spinner.New()
	a.spinner.Spinner = spinner.Globe
	a.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	if a.err != nil {
		return tea.Quit
	}

	if *team == "" {
		a.err = fmt.Errorf("missing team. Use --team to specified the team")
		return tea.Quit
	}

	return tea.Batch(a.spinner.Tick, a.issues.Init())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		a.spinner, cmd = a.spinner.Update(msg)
		if a.loading {
			return a, cmd
		}
		return a, nil
	case loadedIssuesMsg:
		if msg.Done() {
			a.loading = false
			if msg.total == 0 {
				a.err = fmt.Errorf("no issues found")
				return a, tea.Quit
			}
		}
		if msg.total > 0 {
			a.someData = true
		}
	case errMsg:
		a.err = msg
		return a, tea.Quit
	case issueSelectedMsg:
		issue := linear.Issue(msg)
		a.selectedIssue = &issue
		a.quitting = true
		return a, tea.Quit
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			a.quitting = true
			return a, tea.Quit
		case key.Matches(msg, a.keys.Help):
			a.help.ShowAll = !a.help.ShowAll
		default:
		}
	}

	if a.selectedIssue == nil {
		_, cmd = a.issues.Update(msg)
	}
	return a, cmd
}

var style = lipgloss.NewStyle().
	Foreground(lipgloss.Color(linearWhite)).
	Italic(true)

var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D25E6A"))

func (a *App) View() string {
	if a.err != nil {
		return errStyle.Render(a.err.Error()) + "\n"
	}
	if a.selectedIssue != nil && *run == "" {
		i := a.selectedIssue
		return style.Render(fmt.Sprintf("%s %q %s\n", i.Identifier, i.Title, i.BranchName))
	}
	if a.quitting {
		return "\n"
	}
	out := ""
	if a.loading {
		out += a.spinner.View() + "Loading\n"
	}
	if a.someData {
		out += a.issues.View()
	}

	out += a.help.View(a.keys)
	return style.Render(out)
}
