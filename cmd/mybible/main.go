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
var version = "0.3.0"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println("mybible " + version)
		return
	}

	switch flag.Arg(0) {
	case "":
	case "copy-modules":
		if err := copyModules(); err != nil {
			fmt.Fprintln(os.Stderr, "mybible:", err)
			os.Exit(1)
		}
		return
	default:
		fmt.Fprintf(os.Stderr, "mybible: unknown command %q\n", flag.Arg(0))
		flag.Usage()
		os.Exit(2)
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
