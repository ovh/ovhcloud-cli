// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package flags

import (
	"github.com/ovh/ovhcloud-cli/internal/display"
	"gopkg.in/ini.v1"
)

// Flag set from command line
var (
	// Flag to activate debug mode
	Debug bool

	// Flag used to ignore errors in API calls
	IgnoreErrors bool

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

	// Flag to indicate whether the command should use the editor for input parameters
	ParametersViaEditor bool

	// Flag to indicate whether the command should use a file for input parameters
	ParametersFile string
)
