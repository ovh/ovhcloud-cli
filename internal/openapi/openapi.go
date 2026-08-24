// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
func pruneUnknownFields(data map[string]any, schema *openapi3.Schema) map[string]any {
	cleaned := make(map[string]any)
	for propName, propSchema := range schema.Properties {
		if propSchema.Value.ReadOnly {
			continue
		}

		if val, ok := data[propName]; ok {
			// Check if the property is a 'allOf', in which case we take
			// the first schema as the base schema to recurse into
			propValue := propSchema.Value
			if len(propSchema.Value.AllOf) > 0 {
				propValue = propSchema.Value.AllOf[0].Value
			}

			// If the property is an object, recurse
			if propValue.Type.Is("object") {
				// Property is a map of base type, just add it
				if propSchema.Value.AdditionalProperties.Schema != nil {
					cleaned[propName] = val
				} else if nestedMap, ok := val.(map[string]any); ok {
					cleaned[propName] = pruneUnknownFields(nestedMap, propValue)
				} else {
					cleaned[propName] = val
				}
			} else if propValue.Type.Is("array") {
				if val == nil {
					cleaned[propName] = nil
					continue
				}
				arrayVal := val.([]any)
				prunedArray := make([]any, 0, len(arrayVal))
				for _, arrayValue := range arrayVal {
					if arrayMapValue, ok := arrayValue.(map[string]any); ok {
						prunedArray = append(prunedArray, pruneUnknownFields(arrayMapValue, propValue.Items.Value))
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

// GetRequestFieldEnum returns the values a request body field accepts, in the
// order the specification lists them.
//
// Enumerations are worth reading rather than copying: `reason` on a termination
// carries fourteen values today, and a list transcribed into Go drifts the day
// the API gains a fifteenth — silently, into a 400 the operator cannot explain.
// The specification already ships inside the binary, so the values are free.
//
// It returns nil, without an error, for a field that exists but enumerates
// nothing: a caller asking for a free-text field wants an empty list, not a
// failure.
func GetRequestFieldEnum(spec []byte, path, method, field string) ([]string, error) {
	content, err := getRequestBodyFromSpec(spec, path, method)
	if err != nil {
		return nil, err
	}

	if content.Schema == nil || content.Schema.Value == nil {
		return nil, fmt.Errorf("no request schema for %s %s", method, path)
	}

	property, found := content.Schema.Value.Properties[field]
	if !found || property.Value == nil {
		return nil, fmt.Errorf("field %q not found in the body of %s %s", field, method, path)
	}

	values := make([]string, 0, len(property.Value.Enum))
	for _, value := range property.Value.Enum {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("field %q enumerates a %T, expected a string", field, value)
		}
		values = append(values, text)
	}

	return values, nil
}

// GetParameterEnum returns the values a query parameter accepts, in the order
// the specification lists them.
//
// It is the sibling of GetRequestFieldEnum, for the other half of the API
// surface: a read endpoint takes its arguments in the query string, and those
// enumerations are just as long and just as prone to drifting. There are
// fourteen regions and sixteen subsidiaries on the dedicated server scope
// today, and the specification already ships inside the binary.
func GetParameterEnum(spec []byte, path, method, name string) ([]string, error) {
	operation, pathItem, err := getOperationFromSpec(spec, path, method)
	if err != nil {
		return nil, err
	}

	// A parameter may be declared on the operation or shared by every method of
	// the path. Looking at the operation first means a local declaration wins,
	// which is what the specification says happens.
	parameter := operation.Parameters.GetByInAndName("query", name)
	if parameter == nil {
		parameter = pathItem.Parameters.GetByInAndName("query", name)
	}
	if parameter == nil {
		return nil, fmt.Errorf("query parameter %q not found on %s %s", name, method, path)
	}
	if parameter.Schema == nil || parameter.Schema.Value == nil {
		return nil, fmt.Errorf("query parameter %q on %s %s has no schema", name, method, path)
	}

	schema := parameter.Schema.Value
	// A repeatable parameter enumerates its values on the item, not on the array.
	if schema.Type.Is("array") && schema.Items != nil && schema.Items.Value != nil {
		schema = schema.Items.Value
	}

	values := make([]string, 0, len(schema.Enum))
	for _, value := range schema.Enum {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("query parameter %q enumerates a %T, expected a string", name, value)
		}
		values = append(values, text)
	}

	return values, nil
}

// GetComponentEnum returns the values of a named enumeration declared in the
// components of a specification.
//
// Some parameters do not carry their own enumeration: the dedicated server
// availabilities take their datacenters as one comma-separated string, while
// the forty-eight names they accept are declared once, under
// dedicated.AvailabilityDatacenterEnum, and used by the response. Naming the
// component is still better than transcribing forty-eight values into Go, and
// a component that gets renamed fails loudly here rather than quietly offering
// nothing.
func GetComponentEnum(spec []byte, component string) ([]string, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	if doc.Components == nil {
		return nil, fmt.Errorf("specification declares no components")
	}

	schemaRef, found := doc.Components.Schemas[component]
	if !found || schemaRef.Value == nil {
		return nil, fmt.Errorf("component %q not found in spec", component)
	}

	values := make([]string, 0, len(schemaRef.Value.Enum))
	for _, value := range schemaRef.Value.Enum {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("component %q enumerates a %T, expected a string", component, value)
		}
		values = append(values, text)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("component %q enumerates nothing", component)
	}

	return values, nil
}

func getOperationFromSpec(spec []byte, path, method string) (*openapi3.Operation, *openapi3.PathItem, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load spec: %w", err)
	}
	if err = doc.Validate(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("failed to validate spec: %w", err)
	}

	pathItem := doc.Paths.Find(path)
	if pathItem == nil {
		return nil, nil, fmt.Errorf("path %q not found in spec", path)
	}

	operation := pathItem.GetOperation(strings.ToUpper(method))
	if operation == nil {
		return nil, nil, fmt.Errorf("operation %s %s not found", method, path)
	}

	return operation, pathItem, nil
}
