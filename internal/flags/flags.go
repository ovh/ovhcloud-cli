package flags

import (
	"gopkg.in/ini.v1"
	"stash.ovh.net/api/ovh-cli/internal/display"
)

// Flag set from command line
var (
	// Flag to activate debug mode
	Debug bool

	// Common flags used by all subcommands to control output format (json, yaml)
	OutputFormatConfig display.OutputFormat

	// Common filters that can be used in all listing commands
	GenericFilters []string

	// Flag used by all actions that trigger asynchronous tasks to
	// wait for task completion before exiting
	WaitForTask bool

	// INI configuration file and its path
	CliConfig     *ini.File
	CliConfigPath string
)
