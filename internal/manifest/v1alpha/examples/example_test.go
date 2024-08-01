package v1alphaExamples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"testing"

	"github.com/nobl9/nobl9-go/internal/pathutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamples_EnsureAllKindsHaveExamples(t *testing.T) {
	path := filepath.Join(pathutils.FindModuleRoot(), "internal", "manifest", "v1alpha", "examples")
	set := token.NewFileSet()
	packs, err := parser.ParseDir(set, path, nil, 0)
	require.NoError(t, err)

	funcs := []*ast.FuncDecl{}
	for _, pack := range packs {
		for _, f := range pack.Files {
			for _, d := range f.Decls {
				if fn, isFn := d.(*ast.FuncDecl); isFn {
					funcs = append(funcs, fn)
				}
			}
		}
	}

	hasExpectedType := func(f *ast.FuncDecl) bool {
		if f.Type.Params != nil && len(f.Type.Params.List) > 0 {
			return false
		}
		if f.Type.Results == nil || len(f.Type.Results.List) != 1 {
			return false
		}
		at, ok := f.Type.Results.List[0].Type.(*ast.ArrayType)
		if !ok {
			return false
		}
		return at.Elt.(*ast.Ident).Name == "Example"
	}

	for _, kind := range manifest.ApplicableKinds() {
		found := slices.ContainsFunc(funcs, func(f *ast.FuncDecl) bool {
			return kind.String() == f.Name.Name && hasExpectedType(f)
		})
		assert.True(t, found, "missing examples for kind %[1]s, expected function shape: 'func %[1]s() []Example'", kind)
	}
}
