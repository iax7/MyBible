// Command mybible is a terminal reader for MyBible modules.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/tui"
)

// version stays below 1.0 until the module format handling and the UX settle.
// It is a var so release builds can stamp it with -X main.version.
var version = "0.2.2"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("mybible " + version)
		return
	}

	m := tui.New(tui.Options{})
	if err := m.StartupError(); err != nil {
		fmt.Fprintln(os.Stderr, "mybible:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "mybible:", err)
		os.Exit(1)
	}
}
