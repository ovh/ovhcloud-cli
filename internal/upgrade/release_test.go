// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q, want application/json", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v1.4.2","name":"v1.4.2"}`))
	}))
	defer server.Close()

	tag, err := latestTag(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatalf("latestTag: %v", err)
	}
	if tag != "v1.4.2" {
		t.Fatalf("tag = %q, want v1.4.2", tag)
	}
}

func TestLatestTagHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := latestTag(context.Background(), server.Client(), server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLatestTagMissingField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	_, err := latestTag(context.Background(), server.Client(), server.URL)
	if err == nil {
		t.Fatal("expected error for missing tag_name, got nil")
	}
}
