// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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

	t.Run("handles nested object with allOf schema", func(t *testing.T) {
		specWithAllOf := []byte(`
			{
				"openapi": "3.0.0",
				"info": { "title": "Test API", "version": "1.0.0" },
				"components": {
					"schemas": {
						"BaseObj": {
							"type": "object",
							"properties": {
								"baseField": { "type": "string" }
							}
						},
						"ExtendedObj": {
							"allOf": [
								{ "$ref": "#/components/schemas/BaseObj" },
								{
									"type": "object",
									"properties": {
										"extraField": { "type": "integer" }
									}
								}
							]
						}
					}
				},
				"paths": {
					"/allof": {
						"post": {
							"requestBody": {
								"content": {
									"application/json": {
										"schema": {
											"type": "object",
											"properties": {
												"nested": { "$ref": "#/components/schemas/ExtendedObj" }
											}
										},
										"examples": {
											"allOfExample": {
												"value": {
													"nested": {
														"baseField": "base",
														"extraField": 42
													}
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
			"nested": map[string]any{
				"baseField":  "replacedBase",
				"extraField": 99,
			},
		}
		examples, err := GetOperationRequestExamples(specWithAllOf, "/allof", "post", "", replace)
		td.Require(t).CmpNoError(err)
		td.Cmp(t, json.RawMessage(examples["allOfExample"]), td.JSON(`{"nested":{"baseField":"replacedBase","extraField":99}}`))
	})
}
func TestPruneUnknownFields(t *testing.T) {
	// Helper to build openapi3.Schema
	makeSchema := func(props map[string]*openapi3.SchemaRef) *openapi3.Schema {
		return &openapi3.Schema{
			Type:       &openapi3.Types{"object"},
			Properties: props,
		}
	}

	t.Run("removes unknown fields", func(t *testing.T) {
		schema := makeSchema(map[string]*openapi3.SchemaRef{
			"foo": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			"bar": {Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
		})
		input := map[string]any{
			"foo": "abc",
			"bar": 123,
			"baz": "should be removed",
		}
		got := pruneUnknownFields(input, schema)
		td.Cmp(t, got, map[string]any{"foo": "abc", "bar": 123})
	})

	t.Run("skips readOnly fields", func(t *testing.T) {
		schema := makeSchema(map[string]*openapi3.SchemaRef{
			"foo": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, ReadOnly: true}},
			"bar": {Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
		})
		input := map[string]any{"foo": "abc", "bar": 123}
		got := pruneUnknownFields(input, schema)
		td.Cmp(t, got, map[string]any{"bar": 123})
	})

	t.Run("recurses into nested objects", func(t *testing.T) {
		nestedSchema := makeSchema(map[string]*openapi3.SchemaRef{
			"baz": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		})
		schema := makeSchema(map[string]*openapi3.SchemaRef{
			"foo": {Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: nestedSchema.Properties}},
		})
		input := map[string]any{
			"foo":   map[string]any{"baz": "ok", "extra": "remove"},
			"other": "remove",
		}
		got := pruneUnknownFields(input, schema)
		td.Cmp(t, got, map[string]any{"foo": map[string]any{"baz": "ok"}})
	})

	t.Run("handles arrays of objects", func(t *testing.T) {
		itemSchema := &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: map[string]*openapi3.SchemaRef{
			"name": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		}}
		schema := makeSchema(map[string]*openapi3.SchemaRef{
			"items": {Value: &openapi3.Schema{
				Type:  &openapi3.Types{"array"},
				Items: &openapi3.SchemaRef{Value: itemSchema},
			}},
		})
		input := map[string]any{
			"items": []any{
				map[string]any{"name": "a", "extra": "remove"},
				map[string]any{"name": "b"},
			},
		}
		got := pruneUnknownFields(input, schema)
		td.Cmp(t, got, map[string]any{
			"items": []any{
				map[string]any{"name": "a"},
				map[string]any{"name": "b"},
			},
		})
	})

	t.Run("handles nil array value", func(t *testing.T) {
		itemSchema := &openapi3.Schema{Type: &openapi3.Types{"object"}}
		schema := makeSchema(map[string]*openapi3.SchemaRef{
			"items": {Value: &openapi3.Schema{
				Type:  &openapi3.Types{"array"},
				Items: &openapi3.SchemaRef{Value: itemSchema},
			}},
		})
		input := map[string]any{"items": nil}
		got := pruneUnknownFields(input, schema)
		td.Cmp(t, got, map[string]any{"items": nil})
	})

	t.Run("handles allOf schema", func(t *testing.T) {
		base := makeSchema(map[string]*openapi3.SchemaRef{
			"baseField": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		})
		extended := &openapi3.Schema{
			AllOf: []*openapi3.SchemaRef{
				{Value: base},
				{Value: &openapi3.Schema{Properties: map[string]*openapi3.SchemaRef{
					"extraField": {Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
				}}},
			},
		}
		schema := makeSchema(map[string]*openapi3.SchemaRef{
			"nested": {Value: extended},
		})
		input := map[string]any{
			"nested": map[string]any{
				"baseField":  "base",
				"extraField": 42,
				"removeMe":   "no",
			},
		}
		got := pruneUnknownFields(input, schema)
		td.Cmp(t, got, map[string]any{
			"nested": map[string]any{
				"baseField": "base",
			},
		})
	})
}
