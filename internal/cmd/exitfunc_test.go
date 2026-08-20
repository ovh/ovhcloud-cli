// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
	"github.com/ovh/ovhcloud-cli/internal/display"
)

// The suite replaces display.ExitFunc with a no-op (cmd_test.go), so a command
// that calls OutputError or OutputWarning halfway through keeps running under
// test and stops dead in the shipped binary. Every line after such a call is
// dead code that the tests still exercise: whatever those lines print, the
// operator never sees.
//
// executeWithRealExit restores the real semantics for one command. os.Exit
// cannot be called from a test, so the stop is a panic caught here; what
// matters is only that execution does not continue past the call, which is
// what os.Exit does. The output collected is display.ResultString — whatever
// the command managed to produce before it was cut short.
type exitSentinel int

func executeWithRealExit(t *td.T, args ...string) (out string, code int, exited bool) {
	saved := display.ExitFunc
	t.Cleanup(func() { display.ExitFunc = saved })

	defer func() {
		if r := recover(); r != nil {
			stop, ok := r.(exitSentinel)
			if !ok {
				panic(r)
			}
			code, exited, out = int(stop), true, display.ResultString
		}
	}()

	display.ExitFunc = func(c int) { panic(exitSentinel(c)) }
	out, _ = cmd.Execute(args...)
	return out, 0, false
}
