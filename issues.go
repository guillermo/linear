package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guillermo/linear/linear-api"
)

const (
	linearBlue  = "#5E6AD2"
	linearWhite = "#F4F5F8"
	linearBlack = "#222326"
)

type Issues struct {
	err        error
	nextPage   string
	issues     []linear.Issue
	table      table.Model
	termWidth  int
	termHeight int
}

func (t *Issues) Init() tea.Cmd {
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 70},
		{Title: "State", Width: 20},
	}

	t.table = table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(linearBlue)).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(linearWhite)).
		Background(lipgloss.Color(linearBlue)).
		Bold(true)
	t.table.SetStyles(s)

	t.termWidth = 80
	t.termHeight = 24
	return tea.Batch(t.loadIssues)
}

type (
	errMsg error
)

type loadedIssuesMsg struct {
	next  tea.Cmd
	total int
}

type issueSelectedMsg linear.Issue

func (msg loadedIssuesMsg) Done() bool {
	return msg.next == nil
}

func (t *Issues) loadIssues() tea.Msg {
	filter := &linear.IssueFilter{
		Team: &linear.TeamFilter{
			Name: &linear.StringComparator{
				Eq: team,
			},
		},
		State: &linear.WorkflowStateFilter{
			Type: &linear.StringComparator{
				In: []string{"backlog", "unstarted", "started", "completed", "cancelled"},
			},
		},
	}
	var after *string
	if t.nextPage != "" {
		after = &t.nextPage
	}

	res, err := linear.FetchIssues(linear.DefaultClient(), context.Background(), filter, p[int32](30), after)
	if err != nil {
		return errMsg(err)
	}
	if res == nil {
		return errMsg(fmt.Errorf("res is nil"))
	}
	if res.PageInfo == nil {
		return errMsg(fmt.Errorf("linear did not return PageInfo"))
	}

	t.issues = append(t.issues, res.Nodes...)
	sort.Sort(ByState(t.issues))

	msg := loadedIssuesMsg{total: len(t.issues)}

	if res.PageInfo.EndCursor != nil {
		t.nextPage = *res.PageInfo.EndCursor
		msg.next = t.loadIssues
	}

	return msg
}

func issueSelected(i linear.Issue) {
	fmt.Println(i.BranchName)
}

func (t *Issues) setTableSize() {
	padding := 4
	termMargin := 4

	rows := len(t.table.Rows())

	// Default table to the minium
	height := 10 + padding

	// Set the table to the rows
	if rows > height {
		height = rows
	}

	// Set the maxyimumimum
	if height > t.termHeight-termMargin {
		height = t.termHeight - termMargin
	}

	t.table.SetHeight(height)
}

type ByState []linear.Issue

func (a ByState) Len() int      { return len(a) }
func (a ByState) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByState) Less(i, j int) bool {
	if a[i].State.Position == a[j].State.Position {
		return a[i].SortOrder < a[j].SortOrder
	}
	return a[i].State.Position < a[j].State.Position
}

func formatRow(issue linear.Issue) table.Row {
	/*
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(issue.State.Color))

		state := style.Render(issue.State.Name)
	*/

	return table.Row{issue.Identifier, issue.Title, issue.State.Name}
}

func (t *Issues) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.termHeight = msg.Height
		t.termWidth = msg.Width
		t.setTableSize()
	case errMsg:
		t.err = msg
	case loadedIssuesMsg:
		rows := []table.Row{}
		for _, issue := range t.issues {
			rows = append(rows, formatRow(issue))
		}
		t.table.SetRows(rows)
		t.setTableSize()

		if msg.next != nil {
			cmds = append(cmds, msg.next)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			c := t.table.Cursor()
			if c <= len(t.issues) {
				return t, func() tea.Msg { return issueSelectedMsg(t.issues[c]) }
			}
		}
	}
	var tcmd tea.Cmd

	t.table, tcmd = t.table.Update(msg)
	cmds = append(cmds, tcmd)
	return t, tea.Batch(cmds...)
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(linearBlue))

func (t *Issues) View() string {
	return baseStyle.Render(t.table.View()) + "\n"
}
