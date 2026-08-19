// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
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

// A guard that panics has stopped being a guard.
//
// Since #242 every disruptive command asks IsInputFromPipe whether it may
// prompt, so this runs before any confirmation can refuse. os.Stdin.Stat()
// fails on a closed descriptor and returns a nil FileInfo; dereferencing it
// crashes the command with a nil pointer instead of declining the action.
//
// Closing os.Stdin is what the test is about, so it is restored afterwards
// rather than left broken for the rest of the package.
func TestIsInputFromPipeDoesNotPanicOnAClosedStdin(t *testing.T) {
	original := os.Stdin
	t.Cleanup(func() { os.Stdin = original })

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create a pipe: %s", err)
	}
	write.Close()
	read.Close()
	os.Stdin = read

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("IsInputFromPipe panicked on a closed stdin: %v", r)
		}
	}()

	// The value does not matter; not crashing does. A descriptor that cannot be
	// stated is not a terminal, so the guarded command must refuse rather than
	// prompt.
	if IsInputFromPipe() != true {
		t.Error("an unstatable stdin is not an interactive terminal")
	}
}
