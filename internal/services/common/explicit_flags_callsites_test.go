// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// Call sites that legitimately hand over something other than a pointer to a
// package-level parameters struct. Each one is safe for a different reason.
var editResourceValueArgAllowed = map[string]string{
	"nil":          "vrackServices exposes no editable parameter",
	"payload":      "container registry OIDC already builds its body from cmd.Flags().Changed",
	"renewPayload": "service-info already builds its body from cmd.Flags().Changed",
	"userSpec":     "managed database and analytics rebuild a struct per engine, no flag points into it",
}

// addExplicitlySetFlags matches flags to struct fields by address, because a
// flag stores the pointer it was given. A copy of the parameters struct holds
// none of those addresses, so passing one by value turns the whole mechanism
// into a silent no-op for that command — the exact defect this package exists
// to fix. Every call site must therefore hand over a pointer.
func TestEditResourceCallSitesPassAPointer(t *testing.T) {
	servicesDir, err := filepath.Abs("..")
	if err != nil {
		t.Skipf("cannot locate the services directory: %s", err)
	}

	fileSet := token.NewFileSet()
	var checked int

	err = filepath.WalkDir(servicesDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}

		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			return err
		}

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || !isEditResourceCall(call.Fun) {
				return true
			}

			checked++
			if len(call.Args) < 4 {
				t.Errorf("%s: EditResource called with %d arguments, cannot locate the parameters struct",
					fileSet.Position(call.Pos()), len(call.Args))
				return true
			}
			position := fileSet.Position(call.Args[3].Pos())

			switch argument := call.Args[3].(type) {
			case *ast.UnaryExpr:
				if argument.Op != token.AND {
					t.Errorf("%s: parameters passed by value, addExplicitlySetFlags will do nothing", position)
				}
			case *ast.CompositeLit:
				// A body built by hand as a map: every key is present, so
				// omitempty cannot drop anything. A struct literal is another
				// matter — no flag points into a value built on the spot, so
				// the mechanism would silently do nothing.
				if _, isMap := argument.Type.(*ast.MapType); !isMap {
					t.Errorf("%s: a struct literal cannot be matched to any flag, pass a pointer to the parameters struct", position)
				}
			case *ast.Ident:
				if _, allowed := editResourceValueArgAllowed[argument.Name]; !allowed {
					t.Errorf("%s: %s is passed by value, pass &%s so flags set to their zero value survive",
						position, argument.Name, argument.Name)
				}
			default:
				t.Errorf("%s: unexpected parameters argument, expected a pointer", position)
			}

			return true
		})

		return nil
	})

	td.Require(t).CmpNoError(err)
	// Guard against a walk that silently found nothing and passed by default.
	td.Cmp(t, checked > 50, true, "expected the services tree to expose many edit commands, found %d", checked)
}

// isEditResourceCall matches on the function name alone, whatever package
// qualifier carries it: pinning it to "common" would miss an aliased import,
// and a false positive here fails loudly rather than passing in silence.
func isEditResourceCall(fun ast.Expr) bool {
	switch callee := fun.(type) {
	case *ast.SelectorExpr:
		return callee.Sel.Name == "EditResource"
	case *ast.Ident:
		return callee.Name == "EditResource"
	}
	return false
}
