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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// durationPattern is the pattern every metav1.Duration field carries. It must
// match exactly what time.ParseDuration accepts.
const durationPattern = `^[-+]?(0|(([0-9]+(\.[0-9]*)?|\.[0-9]+)(ns|us|µs|μs|ms|s|m|h))+)$`

// A metav1.Duration is a plain string in the CRD. A value time.ParseDuration
// rejects ("30d") passes the API server, then fails to decode in the informer,
// which stops the watch for every object of that kind.
func TestDurationPatternMatchesParseDuration(t *testing.T) {
	re := regexp.MustCompile(durationPattern)
	for _, in := range []string{
		"30d", "1w", "", "h", "1", "1hh", "1 h", "1h ", "--1h", "1.h.5", ".", "1.2.3s",
		"0", "1h", "720h", "1h30m", "1h0m0s", "1.5h", ".5s", "5.s", "-1h", "+1h",
		"100ms", "10us", "10µs", "10μs", "3ns", "2h45m30.5s",
	} {
		_, err := time.ParseDuration(in)
		require.Equal(t, err == nil, re.MatchString(in), "%q", in)
	}
}

// Every metav1.Duration field in the API types carries the pattern marker.
func TestDurationFieldsCarryPattern(t *testing.T) {
	files, err := filepath.Glob("*_types.go")
	require.NoError(t, err)

	marker := "+kubebuilder:validation:Pattern=`" + durationPattern + "`"
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

// The generated CRDs carry the pattern at least once per field (a type used in
// two places appears twice). Run `make manifests` after a change.
func TestCRDsCarryDurationPattern(t *testing.T) {
	files, err := filepath.Glob("../../config/crd/bases/*.yaml")
	require.NoError(t, err)
	found := 0
	for _, f := range files {
		raw, err := os.ReadFile(f)
		require.NoError(t, err)
		found += strings.Count(string(raw), "pattern: "+durationPattern+"\n")
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
