// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package assets

import (
	"context"
	"embed"
	"io/fs"
	"path"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// The whole directory is embedded here, not the *.json inside it, so that a
// file which is not a schema still reaches the test. A refresh that failed
// halfway used to leave one behind, and a leftover nobody sees is a leftover
// somebody commits.
//
//go:embed api-schemas
var schemaFiles embed.FS

// knownInvalidSchemas are the specifications the API publishes in a state the
// OpenAPI validator rejects, with the reason. They are named here rather than
// skipped silently, and the test below insists they still fail: the day one is
// fixed upstream, or the validator stops minding, this list goes red and asks
// to be shortened.
//
// telephony.json was listed until kin-openapi v0.149: two POST operations on
// .../hunting/queue/{queueId}/agent do not declare queueId as a path
// parameter. The schema still does not, but the validator no longer reports
// it, so nothing in the CLI trips on it any more.
var knownInvalidSchemas = map[string]string{}

// TestEmbeddedSchemasAreSchemas reads every embedded specification the way the
// CLI reads it.
//
// Nothing else in the build looks inside these files. `go:embed` is happy with
// an empty one, `go build` is happy with an empty one, and so is `go vet`; the
// first thing that is not happy is a command asking for an enumeration at
// runtime, which answers "value of openapi must be a non-empty string" to
// somebody who typed a server name. Refreshing a schema over a broken network
// used to produce exactly that, silently and with a zero exit code.
func TestEmbeddedSchemasAreSchemas(t *testing.T) {
	entries, err := fs.ReadDir(schemaFiles, "api-schemas")
	if err != nil {
		t.Fatalf("failed to list the embedded schemas: %s", err)
	}

	if len(entries) == 0 {
		t.Fatal("no schema is embedded, which cannot be right")
	}

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			if entry.IsDir() {
				t.Fatalf("%s is a directory; the schemas are flat files", entry.Name())
			}

			// A refresh writes through a temporary file named after the schema
			// with a random suffix. Anything that is not a .json here is either
			// one of those, or something nobody meant to ship.
			if path.Ext(entry.Name()) != ".json" {
				t.Fatalf("%s is not a .json file; a failed schema refresh leaves one of these behind", entry.Name())
			}

			content, err := schemaFiles.ReadFile("api-schemas/" + entry.Name())
			if err != nil {
				t.Fatalf("failed to read %s: %s", entry.Name(), err)
			}

			if len(content) == 0 {
				t.Fatalf("%s is empty", entry.Name())
			}

			doc, err := openapi3.NewLoader().LoadFromData(content)
			if err != nil {
				t.Fatalf("%s does not parse as an OpenAPI document: %s", entry.Name(), err)
			}

			// Parsing alone accepts "{}", and an empty object embeds and builds
			// exactly like a real schema. These two are what a document has to
			// carry for anything to be looked up inside it.
			if doc.OpenAPI == "" {
				t.Fatalf("%s declares no OpenAPI version", entry.Name())
			}
			if doc.Paths == nil || doc.Paths.Len() == 0 {
				t.Fatalf("%s declares no path, so nothing can be looked up in it", entry.Name())
			}

			// The validation internal/openapi runs before it can read anything
			// out of a schema, with the sibling fields it allows next to a $ref.
			// Without them every OVHcloud schema fails, which says nothing
			// about the schema. A document that fails it anyway is recorded in
			// knownInvalidSchemas rather than hidden.
			err = doc.Validate(context.Background(), openapi3.AllowExtraSiblingFields("description", "nullable", "readOnly"))
			reason, known := knownInvalidSchemas[entry.Name()]

			switch {
			case err != nil && !known:
				t.Fatalf("%s does not validate as an OpenAPI document: %s", entry.Name(), err)
			case err == nil && known:
				t.Fatalf("%s validates now (%q was the reason it did not); remove it from knownInvalidSchemas", entry.Name(), reason)
			}
		})
	}
}
