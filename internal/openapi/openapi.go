package openapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
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

func GetOperationRequestExamples(spec []byte, path, method string) (map[string]string, error) {
	content, err := getRequestBodyFromSpec(spec, path, method)
	if err != nil {
		return nil, err
	}

	examples := make(map[string]string, len(content.Examples))

	for k, v := range content.Examples {
		example, err := json.MarshalIndent(v.Value.Value, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshel request body example: %w", err)
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
func pruneUnknownFields(data map[string]interface{}, schema *openapi3.Schema) map[string]interface{} {
	cleaned := make(map[string]interface{})
	for propName, propSchema := range schema.Properties {

		if propSchema.Value.ReadOnly {
			continue
		}

		if val, ok := data[propName]; ok {
			// If the property is an object, recurse
			// TODO: handle arrays
			if propSchema.Value.Type.Is("object") {
				if nestedMap, ok := val.(map[string]interface{}); ok {
					cleaned[propName] = pruneUnknownFields(nestedMap, propSchema.Value)
				} else {
					cleaned[propName] = val
				}
			} else {
				cleaned[propName] = val
			}
		}
	}
	return cleaned
}
