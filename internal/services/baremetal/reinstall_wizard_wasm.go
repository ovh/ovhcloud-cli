// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build js && wasm

package baremetal

import "fmt"

func runReinstallWizard(_ string) (map[string]any, bool, string, error) {
	return nil, false, "", fmt.Errorf("the interactive wizard is not available in this build")
}
