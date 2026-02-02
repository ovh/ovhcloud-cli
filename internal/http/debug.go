// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/flags"
)

// DebugLogEntry represents a single API request/response log entry
type DebugLogEntry struct {
	Timestamp   time.Time
	Method      string
	URL         string
	QueryString string
	StatusCode  int
	RequestID   string
	Duration    time.Duration
	Error       string
}

// DebugLogger holds the debug log entries for the browser
type DebugLogger struct {
	mu      sync.RWMutex
	entries []DebugLogEntry
	maxSize int
}

// Global debug logger instance
var BrowserDebugLogger = &DebugLogger{
	entries: make([]DebugLogEntry, 0),
	maxSize: 100, // Keep last 100 entries
}

// AddEntry adds a new log entry
func (dl *DebugLogger) AddEntry(entry DebugLogEntry) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.entries = append(dl.entries, entry)
	// Keep only the last maxSize entries
	if len(dl.entries) > dl.maxSize {
		dl.entries = dl.entries[len(dl.entries)-dl.maxSize:]
	}
}

// GetEntries returns a copy of all log entries
func (dl *DebugLogger) GetEntries() []DebugLogEntry {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	result := make([]DebugLogEntry, len(dl.entries))
	copy(result, dl.entries)
	return result
}

// Clear removes all log entries
func (dl *DebugLogger) Clear() {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.entries = make([]DebugLogEntry, 0)
}

// Format formats a log entry for display
func (e DebugLogEntry) Format() string {
	status := fmt.Sprintf("%d", e.StatusCode)
	if e.Error != "" {
		status = "ERR"
	}
	reqID := e.RequestID
	if reqID == "" {
		reqID = "-"
	}
	return fmt.Sprintf("[%s] %s %s → %s (%s) RequestID: %s",
		e.Timestamp.Format("15:04:05"),
		e.Method,
		e.URL,
		status,
		e.Duration.Round(time.Millisecond),
		reqID,
	)
}

type transport struct {
	name      string
	transport http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	if flags.Debug {
		reqData, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Printf("[DEBUG] "+logReqMsg, t.name, prettyPrintJsonLines(reqData))
		} else {
			log.Printf("[ERROR] %s API Request error: %#v", t.name, err)
		}
	}

	resp, err := t.transport.RoundTrip(req)
	duration := time.Since(startTime)

	// Create debug log entry
	entry := DebugLogEntry{
		Timestamp:   startTime,
		Method:      req.Method,
		URL:         req.URL.Path,
		QueryString: req.URL.RawQuery,
		Duration:    duration,
	}

	if err != nil {
		entry.Error = err.Error()
		BrowserDebugLogger.AddEntry(entry)
		return resp, err
	}

	entry.StatusCode = resp.StatusCode
	// Extract X-OVH-QUERYID from response headers
	entry.RequestID = resp.Header.Get("X-OVH-QUERYID")

	BrowserDebugLogger.AddEntry(entry)

	if flags.Debug {
		respData, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Printf("[DEBUG] "+logRespMsg, t.name, prettyPrintJsonLines(respData))
		} else {
			log.Printf("[ERROR] %s API Response error: %#v", t.name, err)
		}
	}

	return resp, nil
}

// NewTransport creates a wrapper around a *http.RoundTripper,
// designed to be used for the `Transport` field of http.Client.
//
// This logs each pair of HTTP request/response that it handles.
// The logging is done via Go standard library `log` package.
//
// Deprecated: This will log the content of every http request/response
// at `[DEBUG]` level, without any filtering. Any sensitive information
// will appear as-is in your logs. Please use NewSubsystemLoggingHTTPTransport instead.
func NewTransport(name string, t http.RoundTripper) *transport {
	return &transport{name, t}
}

// prettyPrintJsonLines iterates through a []byte line-by-line,
// transforming any lines that are complete json into pretty-printed json.
func prettyPrintJsonLines(b []byte) string {
	parts := strings.Split(string(b), "\n")
	for i, p := range parts {
		if b := []byte(p); json.Valid(b) {
			var out bytes.Buffer
			_ = json.Indent(&out, b, "", " ") // already checked for validity
			parts[i] = out.String()
		}
	}
	return strings.Join(parts, "\n")
}

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`
