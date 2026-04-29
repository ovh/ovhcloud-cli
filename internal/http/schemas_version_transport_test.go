// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"
	"testing"

	td "github.com/maxatome/go-testdeep"
)

type recordingTransport struct {
	req *http.Request
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func TestSchemasVersionTransport_V2Path(t *testing.T) {
	recorder := &recordingTransport{}
	transport := newSchemasVersionTransport(recorder)

	req, _ := http.NewRequest(http.MethodGet, "https://eu.api.ovh.com/v2/vrackServices/resource/abc", nil)
	_, err := transport.RoundTrip(req)
	td.Require(t).CmpNoError(err)
	td.Cmp(t, recorder.req.Header.Get(schemasVersionHeader), schemasVersion)
}

func TestSchemasVersionTransport_V1Path(t *testing.T) {
	recorder := &recordingTransport{}
	transport := newSchemasVersionTransport(recorder)

	req, _ := http.NewRequest(http.MethodGet, "https://eu.api.ovh.com/v1/dedicated/server/abc", nil)
	_, err := transport.RoundTrip(req)
	td.Require(t).CmpNoError(err)
	td.Cmp(t, recorder.req.Header.Get(schemasVersionHeader), "")
}

func TestSchemasVersionTransport_NoPrefix(t *testing.T) {
	recorder := &recordingTransport{}
	transport := newSchemasVersionTransport(recorder)

	req, _ := http.NewRequest(http.MethodGet, "https://eu.api.ovh.com/dedicated/server/abc", nil)
	_, err := transport.RoundTrip(req)
	td.Require(t).CmpNoError(err)
	td.Cmp(t, recorder.req.Header.Get(schemasVersionHeader), "")
}

func TestSchemasVersionTransport_DoesNotMutateOriginal(t *testing.T) {
	recorder := &recordingTransport{}
	transport := newSchemasVersionTransport(recorder)

	req, _ := http.NewRequest(http.MethodGet, "https://eu.api.ovh.com/v2/vrackServices/resource/abc", nil)
	_, err := transport.RoundTrip(req)
	td.Require(t).CmpNoError(err)
	// The original request must not have been mutated.
	td.Cmp(t, req.Header.Get(schemasVersionHeader), "")
}
