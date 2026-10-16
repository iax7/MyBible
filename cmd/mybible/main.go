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
const version = "0.0.2"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("mybible " + version)
		return
	}

	if _, err := tea.NewProgram(tui.New(tui.Options{})).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "mybible:", err)
		os.Exit(1)
	}
}
