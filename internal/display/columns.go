// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package display

import (
	"regexp"
	"strings"
)

// dottedColumnPath matches a column that is a plain path through nested
// objects, such as "currentState.status".
var dottedColumnPath = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)+$`)

// columnParentIsMissing reports whether the given column is a path through an
// object that is absent or null in the row, e.g. "currentState.status" while an
// API v2 resource is CREATING. Such a column is an empty cell for that row.
//
// The absent case is the usual one, and cannot be told apart from a mistyped
// parent: API v2 omits an empty currentState instead of sending null.
//
// It returns false when a parent of the path exists but is not an object: that
// is a wrong column definition, which must keep failing.
func columnParentIsMissing(row map[string]any, column string) bool {
	if !dottedColumnPath.MatchString(column) {
		return false
	}

	keys := strings.Split(column, ".")
	var current any = row
	for _, key := range keys[:len(keys)-1] {
		object, ok := current.(map[string]any)
		if !ok {
			return false
		}
		current = object[key]
		if current == nil {
			return true
		}
	}

	return false
}
