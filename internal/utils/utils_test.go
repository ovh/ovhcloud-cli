// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestMergeMaps_BasicMerge(t *testing.T) {
	left := map[string]any{
		"a": 1,
		"b": "foo",
	}
	right := map[string]any{
		"b": "bar",
		"c": 3,
	}
	expected := map[string]any{
		"a": 1,
		"b": "bar",
		"c": 3,
	}

	err := MergeMaps(left, right)

	td.CmpNoError(t, err)
	td.Cmp(t, left, expected)
}

func TestMergeMaps_EmptyRight(t *testing.T) {
	left := map[string]any{"a": 1}
	right := map[string]any{}
	expected := map[string]any{"a": 1}

	err := MergeMaps(left, right)
	td.CmpNoError(t, err)
	td.Cmp(t, left, expected)
}

func TestMergeMaps_EmptyLeft(t *testing.T) {
	left := map[string]any{}
	right := map[string]any{"a": 2}
	expected := map[string]any{"a": 2}

	err := MergeMaps(left, right)
	td.CmpNoError(t, err)
	td.Cmp(t, left, expected)
}

func TestMergeMaps_SliceAppend(t *testing.T) {
	left := map[string]any{"arr": []int{1, 2}}
	right := map[string]any{"arr": []int{3, 4}}
	expected := map[string]any{"arr": []int{1, 2, 3, 4}}

	err := MergeMaps(left, right)
	td.CmpNoError(t, err)
	td.Cmp(t, left, expected)
}

func TestMergeMaps_WrongType(t *testing.T) {
	left := map[string]any{"a": "1"}
	right := map[string]any{"a": 1}
	expected := map[string]any{"a": 1}

	err := MergeMaps(left, right)
	td.CmpNoError(t, err)
	td.Cmp(t, left, expected)
}
