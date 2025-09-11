// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ovh/ovhcloud-cli/internal/utils"
)

func FilterEditableFields(spec []byte, path, method string, body map[string]any) (map[string]any, error) {
	content, err := getRequestBodyFromSpec(spec, path, method)
	if err != nil {
		return nil, err
	}

	// Prune unknown fields
	pruned := pruneUnknownFields(body, content.Schema.Value)

	return pruned, nil
}

func GetOperationRequestExamples(spec []byte, path, method, defaultExample string, replaceValues map[string]any) (map[string]string, error) {
	content, err := getRequestBodyFromSpec(spec, path, method)
	if err != nil {
		return nil, err
	}

	jsonExamples := make(map[string][]byte, len(content.Examples)+1)
	for k, v := range content.Examples {
		// Marshal & unmarshal example to get the request
		// body example as a map[string]any
		jsonExample, err := json.Marshal(v.Value.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request example: %s", err)
		}
		jsonExamples[k] = jsonExample
	}
	if defaultExample != "" {
		jsonExamples["default"] = []byte(defaultExample)
	}

	examples := make(map[string]string, len(content.Examples))

	for k, v := range jsonExamples {
		var objectExample map[string]any
		if err := json.Unmarshal(v, &objectExample); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request example: %s", err)
		}

		// Merge replace values with the example
		if err := utils.MergeMaps(objectExample, replaceValues); err != nil {
			return nil, fmt.Errorf("failed to merge replace values into example: %w", err)
		}

		// Marshal the final merged example
		example, err := json.MarshalIndent(objectExample, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body example: %w", err)
		}

		examples[k] = string(example)
	}

	return examples, nil
}

func getRequestBodyFromSpec(spec []byte, path, method string) (*openapi3.MediaType, error) {
	// Load the OpenAPI spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}
	if err = doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to validate spec: %w", err)
	}

	// Retrieve operation
	var (
		pathItem = doc.Paths.Find(path)
		op       *openapi3.Operation
	)

	if pathItem == nil {
		return nil, fmt.Errorf("path %q not found in spec", path)
	}

	switch method {
	case "put":
		op = pathItem.Put
	case "post":
		op = pathItem.Post
	default:
		return nil, fmt.Errorf("invalid write method %q", method)
	}

	if op == nil {
		return nil, fmt.Errorf("operation %s %s not found", method, path)
	}

	// Get request body
	reqBody := op.RequestBody.Value

	return reqBody.Content["application/json"], nil
}

// pruneUnknownFields recursively removes fields not in the schema
func pruneUnknownFields(data map[string]any, schema *openapi3.Schema) map[string]interface{} {
	cleaned := make(map[string]any)
	for propName, propSchema := range schema.Properties {

		if propSchema.Value.ReadOnly {
			continue
		}

		if val, ok := data[propName]; ok {
			// If the property is an object, recurse
			if propSchema.Value.Type.Is("object") {
				// Property is a map of base type, just add it
				if propSchema.Value.AdditionalProperties.Schema != nil {
					cleaned[propName] = val
				} else if nestedMap, ok := val.(map[string]any); ok {
					cleaned[propName] = pruneUnknownFields(nestedMap, propSchema.Value)
				} else {
					cleaned[propName] = val
				}
			} else if propSchema.Value.Type.Is("array") {
				if val == nil {
					cleaned[propName] = nil
					continue
				}
				arrayVal := val.([]any)
				prunedArray := make([]any, 0, len(arrayVal))
				for _, arrayValue := range arrayVal {
					if arrayMapValue, ok := arrayValue.(map[string]any); ok {
						prunedArray = append(prunedArray, pruneUnknownFields(arrayMapValue, propSchema.Value.Items.Value))
					} else {
						prunedArray = append(prunedArray, arrayValue)
					}
				}
				cleaned[propName] = prunedArray
			} else {
				cleaned[propName] = val
			}
		}
	}
	return cleaned
}
