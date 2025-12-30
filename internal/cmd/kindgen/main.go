package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed kind.tmpl
var templateStr string

type Template struct {
	ProgramInvocation  string
	ProjectScopedKinds []string
}

func main() {
	programName := filepath.Base(os.Args[0])
	fmt.Printf("%s: scanning for ProjectScopedObject implementations\n", programName)

	cwd, err := os.Getwd()
	if err != nil {
		errFatal(err.Error())
	}

	// When run via go generate from the manifest package, cwd is the manifest directory.
	// v1alpha is a subdirectory of manifest.
	v1alphaDir := filepath.Join(cwd, "v1alpha")
	projectScopedKinds, err := findProjectScopedKinds(v1alphaDir)
	if err != nil {
		errFatal(err.Error())
	}

	fmt.Printf("%s: found %d project-scoped kinds: %v\n", programName, len(projectScopedKinds), projectScopedKinds)

	tpl, err := template.New("generator").Parse(templateStr)
	if err != nil {
		errFatal(err.Error())
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, Template{
		ProgramInvocation:  programName,
		ProjectScopedKinds: projectScopedKinds,
	})
	if err != nil {
		errFatal(err.Error())
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		errFatal(fmt.Sprintf("failed to format generated code: %s\n%s", err.Error(), buf.String()))
	}

	outputPath := filepath.Join(cwd, "kind_project_scoped.go")
	if err = os.WriteFile(outputPath, formatted, 0o600); err != nil {
		errFatal(err.Error())
	}

	fmt.Printf("%s: generated %s\n", programName, outputPath)
}

func findProjectScopedKinds(v1alphaDir string) ([]string, error) {
	var kinds []string

	err := filepath.WalkDir(v1alphaDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, "_object.go") {
			return nil
		}

		typeName, isProjectScoped, err := checkProjectScopedObject(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}
		if isProjectScoped {
			kinds = append(kinds, "Kind"+typeName)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(kinds)
	return kinds, nil
}

func checkProjectScopedObject(filePath string) (typeName string, isProjectScoped bool, err error) {
	fileset := token.NewFileSet()
	f, err := parser.ParseFile(fileset, filePath, nil, 0)
	if err != nil {
		return "", false, err
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) != 1 {
				continue
			}
			// Check for: var _ manifest.ProjectScopedObject = TypeName{}
			if valueSpec.Names[0].Name != "_" {
				continue
			}
			selExpr, ok := valueSpec.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			ident, ok := selExpr.X.(*ast.Ident)
			if !ok || ident.Name != "manifest" || selExpr.Sel.Name != "ProjectScopedObject" {
				continue
			}
			// Found it, now extract the type name from the value
			if len(valueSpec.Values) != 1 {
				continue
			}
			compositeLit, ok := valueSpec.Values[0].(*ast.CompositeLit)
			if !ok {
				continue
			}
			typeIdent, ok := compositeLit.Type.(*ast.Ident)
			if !ok {
				continue
			}
			return typeIdent.Name, true, nil
		}
	}

	return "", false, nil
}

func errFatal(f string) {
	fmt.Fprintln(os.Stderr, f)
	os.Exit(1)
}
