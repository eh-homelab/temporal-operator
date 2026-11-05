// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1beta1_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// durationRule is the CEL rule every metav1.Duration field carries. CEL's
// duration() calls time.ParseDuration, so the API server accepts exactly what
// the operator can decode. A value it rejects ("30d") would otherwise pass the
// API server, then fail to decode in the informer and stop the watch for every
// object of that kind.
const durationRule = `duration(self) >= duration('0s')`

// Every metav1.Duration field in the API types carries the rule marker.
func TestDurationFieldsCarryRule(t *testing.T) {
	files, err := filepath.Glob("*_types.go")
	require.NoError(t, err)

	marker := `+kubebuilder:validation:XValidation:rule="` + durationRule + `"`
	fields := 0
	for _, f := range files {
		file, err := parser.ParseFile(token.NewFileSet(), f, nil, parser.ParseComments)
		require.NoError(t, err)
		ast.Inspect(file, func(n ast.Node) bool {
			field, ok := n.(*ast.Field)
			if !ok || !isDuration(field.Type) {
				return true
			}
			fields++
			require.NotNil(t, field.Doc, "%s: %s has no doc", f, field.Names[0].Name)
			require.Contains(t, field.Doc.Text()+rawComments(field.Doc), marker, "%s: %s", f, field.Names[0].Name)
			return true
		})
	}
	require.Positive(t, fields)
}

// The generated CRDs carry the rule at least once per field (a type used in
// two places appears twice). Run `make manifests` after a change.
func TestCRDsCarryDurationRule(t *testing.T) {
	files, err := filepath.Glob("../../config/crd/bases/*.yaml")
	require.NoError(t, err)
	found := 0
	for _, f := range files {
		raw, err := os.ReadFile(f)
		require.NoError(t, err)
		found += strings.Count(string(raw), "rule: "+durationRule+"\n")
	}
	require.GreaterOrEqual(t, found, countDurationFields(t))
}

func isDuration(expr ast.Expr) bool {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "metav1" && sel.Sel.Name == "Duration"
}

// rawComments returns the comment lines as written; Doc.Text drops "+marker" lines.
func rawComments(g *ast.CommentGroup) string {
	var b strings.Builder
	for _, c := range g.List {
		b.WriteString(c.Text)
		b.WriteString("\n")
	}
	return b.String()
}

func countDurationFields(t *testing.T) int {
	files, err := filepath.Glob("*_types.go")
	require.NoError(t, err)
	n := 0
	for _, f := range files {
		file, err := parser.ParseFile(token.NewFileSet(), f, nil, 0)
		require.NoError(t, err)
		ast.Inspect(file, func(node ast.Node) bool {
			if field, ok := node.(*ast.Field); ok && isDuration(field.Type) {
				n++
			}
			return true
		})
	}
	return n
}
