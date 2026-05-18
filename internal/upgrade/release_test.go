// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestLatestTag(t *testing.T) {
	httpmock.Activate(t)
	httpmock.RegisterResponder("GET", latestReleaseURL,
		httpmock.NewJsonResponderOrPanic(http.StatusOK, json.RawMessage(`{"tag_name":"v1.4.2","name":"v1.4.2"}`)))
	httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Fatal))

	tag, err := LatestTag(context.Background())
	if err != nil {
		t.Fatalf("LatestTag: %v", err)
	}
	if tag != "v1.4.2" {
		t.Fatalf("tag = %q, want v1.4.2", tag)
	}
}

func TestLatestTagHTTPError(t *testing.T) {
	httpmock.Activate(t)
	httpmock.RegisterResponder("GET", latestReleaseURL,
		httpmock.NewStringResponder(http.StatusInternalServerError, "nope"))
	httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Fatal))

	_, err := LatestTag(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLatestTagMissingField(t *testing.T) {
	httpmock.Activate(t)
	httpmock.RegisterResponder("GET", latestReleaseURL,
		httpmock.NewStringResponder(http.StatusOK, `{}`))
	httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Fatal))

	_, err := LatestTag(context.Background())
	if err == nil {
		t.Fatal("expected error for missing tag_name, got nil")
	}
}
