// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// addExplicitlySetFlags puts back into the request body the fields the user
// explicitly set on the command line but whose value was dropped by
// `omitempty` during marshalling — a boolean set to false, an integer set to
// zero, a string set to the empty value.
//
// Without this, `--monitoring=false` produced an empty body, the merge with
// the fetched resource restored the remote value, and the CLI reported a
// success for an update that never happened.
//
// Flags are matched to struct fields by address, not by name: pflag stores the
// pointer it was given (`type boolValue bool` + `newBoolValue(val, p)` returns
// `(*boolValue)(p)`), so the address of a flag value is the address of the
// field it writes to. Matching by name would be wrong — several commands
// expose a field under an unrelated flag name, for instance
// PrivateNetworkRoutingAsDefault as `--routing-as-default`.
func addExplicitlySetFlags(cmd *cobra.Command, cliParams any, body map[string]any) {
	if cmd == nil || cliParams == nil || body == nil {
		return
	}

	value := reflect.ValueOf(cliParams)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return
	}

	// Address of every addressable scalar field, with the path of JSON keys
	// leading to it and the field itself.
	fields := map[uintptr]targetField{}
	collectFieldAddresses(value, nil, fields)

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		flagValue := reflect.ValueOf(flag.Value)
		if flagValue.Kind() != reflect.Ptr || flagValue.IsNil() {
			return
		}

		target, ok := fields[flagValue.Pointer()]
		if !ok {
			return
		}

		// Read the value from the struct field, not from the flag: pflag
		// stores it under its own type (pflag.boolValue), which would end up
		// in the request body instead of a plain bool.
		setAtPath(body, target.path, target.value.Interface())
	})
}

// targetField is a scalar field of the parameters struct: where it belongs in
// the request body, and the field itself.
type targetField struct {
	path  []string
	value reflect.Value
}

// collectFieldAddresses walks the struct and records, for each scalar field,
// its address and the chain of JSON keys leading to it.
func collectFieldAddresses(value reflect.Value, prefix []string, out map[uintptr]targetField) {
	structType := value.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}

		fieldValue := value.Field(i)
		path := append(append([]string{}, prefix...), name)

		if fieldValue.Kind() == reflect.Struct {
			collectFieldAddresses(fieldValue, path, out)
			continue
		}

		if fieldValue.CanAddr() {
			out[fieldValue.Addr().Pointer()] = targetField{path: path, value: fieldValue}
		}
	}
}

// setAtPath writes value in body, creating intermediate maps as needed.
func setAtPath(body map[string]any, path []string, value any) {
	for _, key := range path[:len(path)-1] {
		next, ok := body[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			body[key] = next
		}
		body = next
	}

	body[path[len(path)-1]] = value
}
