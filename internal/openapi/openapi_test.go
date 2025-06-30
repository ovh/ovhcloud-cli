package openapi

import (
	"encoding/json"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestGetOperationRequestExamples(t *testing.T) {
	spec := []byte(`{
	  "openapi": "3.0.0",
	  "info": { "title": "Test API", "version": "1.0.0" },
	  "paths": {
		"/test": {
		  "post": {
			"requestBody": {
			  "content": {
				"application/json": {
				  "schema": {
					"type": "object",
					"properties": {
					  "foo": { "type": "string" },
					  "bar": { "type": "integer" }
					}
				  },
				  "examples": {
					"default": {
					  "value": {
						"foo": "original",
						"bar": 123
					  }
					}
				  }
				}
			  }
			},
			"responses": {
			  "200": {
				"description": "OK"
			  }
			}
		  }
		}
	  }
	}`)

	t.Run("returns merged example with replaceValues", func(t *testing.T) {
		require := td.Require(t)

		replace := map[string]any{"foo": "replaced"}
		examples, err := GetOperationRequestExamples(spec, "/test", "post", "", replace)
		require.CmpNoError(err)
		require.ContainsKey(examples, "default")
		require.Cmp(json.RawMessage(examples["default"]), td.JSON(`{"foo":"replaced","bar":123}`))
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		_, err := GetOperationRequestExamples(spec, "/notfound", "post", "", nil)
		td.CmpString(t, err, `path "/notfound" not found in spec`)
	})

	t.Run("returns error for invalid method", func(t *testing.T) {
		_, err := GetOperationRequestExamples(spec, "/test", "put", "", nil)
		td.CmpString(t, err, "operation put /test not found")
	})

	t.Run("returns error for invalid spec", func(t *testing.T) {
		_, err := GetOperationRequestExamples([]byte("not json"), "/test", "post", "", nil)
		td.CmpContains(t, err, "cannot unmarshal string into Go value of type openapi3.TBis")
	})

	t.Run("returns no error if example is not marshalable", func(t *testing.T) {
		// Spec with an example that cannot be marshaled (function value)
		specWithBadExample := []byte(`{
		  "openapi": "3.0.0",
		  "info": { "title": "Test API", "version": "1.0.0" },
		  "paths": {
			"/test": {
			  "post": {
				"requestBody": {
				  "content": {
					"application/json": {
					  "schema": { "type": "object" },
					  "examples": {
						"bad": { "value": { "$$bad": {} } }
					  }
					}
				  }
				},
				"responses": {
				  "200": {
					"description": "OK"
				  }
				}
			  }
			}
		  }
		}`)

		examples, err := GetOperationRequestExamples(specWithBadExample, "/test", "post", "", nil)
		td.CmpNoError(t, err)
		td.CmpContainsKey(t, examples, "bad")
	})
}
