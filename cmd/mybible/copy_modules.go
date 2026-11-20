package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/mtp"
)

func usage() {
	fmt.Fprint(flag.CommandLine.Output(), `Usage:
  mybible               read, in the terminal
  mybible copy-modules  copy Modules from a phone connected over MTP

Flags:
`)
	flag.PrintDefaults()
}

// copyModules copies the phone's Modules into the folder the reader scans:
// module_directory from the config, or the default when config names none.
func copyModules() error {
	configDir, err := config.Dir()
	if err != nil {
		return err
	}
	cfg, err := config.Load(configDir)
	if err != nil {
		return err
	}
	dest := cfg.ModuleDir
	if dest == "" {
		dest, err = config.DefaultModuleDir()
		if err != nil {
			return err
		}
	}
	return mtp.CopyModules(dest, os.Stdout)
}
