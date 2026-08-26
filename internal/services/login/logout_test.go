// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
)

func TestIsInvalidCredentialError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"forbidden (403)", &ovh.APIError{Code: http.StatusForbidden}, true},
		{"unauthorized (401)", &ovh.APIError{Code: http.StatusUnauthorized}, true},
		{"other API error (500)", &ovh.APIError{Code: http.StatusInternalServerError}, false},
		{"not found (404)", &ovh.APIError{Code: http.StatusNotFound}, false},
		{"wrapped forbidden", fmt.Errorf("revoke failed: %w", &ovh.APIError{Code: http.StatusForbidden}), true},
		{"plain error", errors.New("network unreachable"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td.Cmp(t, isInvalidCredentialError(tc.err), tc.expected)
		})
	}
}
