package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

func p[T any](v T) *T {
	return &v
}

var (
	team      = flag.String("team", "", "linear team to display")
	run       = flag.String("run", "", "execute the program on the selected issue.\nThe programm will be called as:\n\tprg ID \"TITLE\" \"BRANCH\"")
	wantsHelp = flag.Bool("help", false, "display help")
)

func main() {
	flag.Parse()
	if *wantsHelp {
		flag.PrintDefaults()
		return
	}
	app := &App{}
	if os.Getenv("LINEAR_KEY") == "" {
		app.err = errors.New("LINEAR_KEY env var must be set")
	}
	model, err := tea.NewProgram(app).Run()
	if err != nil {
		log.Fatal(err)
	}
	app = model.(*App)

	if app.selectedIssue == nil {
		return
	}

	if *run == "" {
		return
	}

	args := strings.Split(*run, " ")
	if len(args) == 0 {
		return
	}
	prg := args[0]

	binary, lookErr := exec.LookPath(prg)
	if lookErr != nil {
		log.Fatal(lookErr)
	}

	env := os.Environ()
	i := app.selectedIssue
	args = append(args, i.Identifier, i.Title, i.BranchName)

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		log.Fatal(execErr)
	}
}
