// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package apicall

import "os"

// readFile is a variable so a test can supply a payload without writing to disk.
var readFile = os.ReadFile
