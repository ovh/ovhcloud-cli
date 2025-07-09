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

	t.Run("handles request body with map[string]string property", func(t *testing.T) {
		specWithMap := []byte(`
			{
				"openapi": "3.0.0",
				"info": { "title": "Test API", "version": "1.0.0" },
				"paths": {
					"/maptest": {
						"post": {
							"requestBody": {
								"content": {
									"application/json": {
										"schema": {
											"type": "object",
											"properties": {
												"labels": {
													"type": "object",
													"additionalProperties": { "type": "string" }
												},
												"other": { "type": "string" }
											}
										},
										"examples": {
											"mapExample": {
												"value": {
													"labels": { "foo": "bar", "baz": "qux" },
													"other": "value"
												}
											}
										}
									}
								}
							},
							"responses": {
								"200": { "description": "OK" }
							}
						}
					}
				}
			}`,
		)

		replace := map[string]any{
			"labels": map[string]any{"foo": "replaced"},
			"other":  "changed",
		}
		examples, err := GetOperationRequestExamples(specWithMap, "/maptest", "post", "", replace)
		td.Require(t).CmpNoError(err)
		td.Cmp(t, json.RawMessage(examples["mapExample"]), td.JSON(`{"labels":{"foo":"replaced", "baz": "qux"},"other":"changed"}`))
	})

	t.Run("handles request body with map[string]$ref property", func(t *testing.T) {
		specWithMapRef := []byte(`
			{
				"openapi": "3.0.0",
				"info": { "title": "Test API", "version": "1.0.0" },
				"components": {
					"schemas": {
						"SubObj": {
							"type": "object",
							"properties": {
								"subfield": { "type": "string" }
							}
						}
					}
				},
				"paths": {
					"/mapref": {
						"post": {
							"requestBody": {
								"content": {
									"application/json": {
										"schema": {
											"type": "object",
											"properties": {
												"refs": {
													"type": "object",
													"additionalProperties": { "$ref": "#/components/schemas/SubObj" }
												},
												"other": { "type": "string" }
											}
										},
										"examples": {
											"mapRefExample": {
												"value": {
													"refs": { "a": { "subfield": "x" }, "b": { "subfield": "y" } },
													"other": "value"
												}
											}
										}
									}
								}
							},
							"responses": {
								"200": { "description": "OK" }
							}
						}
					}
				}
			}`,
		)

		replace := map[string]any{
			"refs":  map[string]any{"a": map[string]any{"subfield": "replaced"}},
			"other": "changed",
		}
		examples, err := GetOperationRequestExamples(specWithMapRef, "/mapref", "post", "", replace)
		td.Require(t).CmpNoError(err)
		td.Cmp(t, json.RawMessage(examples["mapRefExample"]), td.JSON(`{"refs":{"a":{"subfield":"replaced"}, "b": { "subfield": "y" }},"other":"changed"}`))
	})

	t.Run("uses defaultExample if provided", func(t *testing.T) {
		spec := []byte(`
			{
				"openapi": "3.0.0",
				"info": { "title": "Test API", "version": "1.0.0" },
				"paths": {
					"/default": {
						"post": {
							"requestBody": {
								"content": {
									"application/json": {
										"schema": {
											"type": "object",
											"properties": {
												"foo": { "type": "string" }
											}
										}
									}
								}
							},
							"responses": {
								"200": { "description": "OK" }
							}
						}
					}
				}
			}`,
		)
		defaultExample := `{"foo":"fromDefault"}`
		examples, err := GetOperationRequestExamples(spec, "/default", "post", defaultExample, nil)
		td.Require(t).CmpNoError(err)
		td.Require(t).ContainsKey(examples, "default")
		td.Cmp(t, json.RawMessage(examples["default"]), td.JSON(`{"foo":"fromDefault"}`))
	})
}
