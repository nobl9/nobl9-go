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
	"strings"
)

//go:embed generate-object-impl.tpl
var templateStr string

type generator struct {
	StructName             string
	Receiver               string
	IsProjectScopedObject  bool
	IsV1alphaObjectContext bool
}

type Template struct {
	Receiver          string
	StructName        string
	Package           string
	ProgramInvocation string

	GenerateObject               bool
	GenerateProjectScopedObject  bool
	GenerateV1alphaObjectContext bool
}

func main() {
	if len(os.Args) != 2 {
		errFatal("you must provide struct name")
	}
	g := &generator{StructName: os.Args[1]}

	filename := os.Getenv("GOFILE")
	pkg := os.Getenv("GOPACKAGE")
	programName := filepath.Base(os.Args[0])
	fmt.Printf("%s [Struct: %s, File: %s, Package: %s]\n", programName, g.StructName, filename, pkg)

	cwd, err := os.Getwd()
	if err != nil {
		errFatal(err.Error())
	}

	fst := token.NewFileSet()

	af, err := parser.ParseFile(fst, filepath.Join(cwd, filename), nil, 0)
	if err != nil {
		errFatal(err.Error())
	}

	ast.Inspect(af, g.genDecl)

	tpl, err := template.New("generator").Parse(templateStr)
	if err != nil {
		errFatal(err.Error())
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, Template{
		Receiver:                     string(strings.ToLower(g.StructName)[0]),
		StructName:                   g.StructName,
		Package:                      pkg,
		ProgramInvocation:            fmt.Sprintf("%s %s", programName, strings.Join(os.Args[1:], " ")),
		GenerateObject:               true,
		GenerateProjectScopedObject:  g.IsProjectScopedObject,
		GenerateV1alphaObjectContext: g.IsV1alphaObjectContext && strings.Contains(cwd, "v1alpha"),
	})
	if err != nil {
		errFatal(err.Error())
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		errFatal(err.Error())
	}
	outputName := filepath.Join(cwd, fmt.Sprintf("%s_object.go", strings.TrimSuffix(filename, ".go")))
	if err = os.WriteFile(outputName, formatted, 0644); err != nil {
		errFatal(err.Error())
	}
}

func (g *generator) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		// We only care about type declarations.
		return true
	}
	if len(decl.Specs) != 1 {
		return false
	}
	spec, ok := decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return false
	}
	structType, isStruct := spec.Type.(*ast.StructType)
	if !isStruct || spec.Name.Name != g.StructName {
		return false
	}
	g.IsProjectScopedObject = g.hasProjectInMetadata(structType.Fields)
	g.IsV1alphaObjectContext = g.hasOrganizationAndManifestSource(structType.Fields)
	return false
}

func (g *generator) hasProjectInMetadata(fields *ast.FieldList) bool {
	hasProjectRef := false
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			continue
		}
		if field.Names[0].Name != "Metadata" {
			continue
		}
		metadata, ok := field.Type.(*ast.Ident).
			Obj.Decl.(*ast.TypeSpec).
			Type.(*ast.StructType)
		if !ok {
			continue
		}
		for _, mf := range metadata.Fields.List {
			if len(mf.Names) == 0 {
				continue
			}
			if mf.Names[0].Name == "Project" {
				hasProjectRef = true
				break
			}
		}
	}
	return hasProjectRef
}

func (g *generator) hasOrganizationAndManifestSource(fields *ast.FieldList) bool {
	hasOrganization := false
	hasManifestSource := false
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			continue
		}
		if field.Names[0].Name == "Organization" {
			hasOrganization = true
			continue
		}
		if field.Names[0].Name == "ManifestSource" {
			hasManifestSource = true
			continue
		}
	}
	return hasOrganization && hasManifestSource
}

func errFatal(f string, a ...interface{}) {
	if len(a) == 0 {
		fmt.Fprintln(os.Stderr, f)
	} else {
		fmt.Fprintf(os.Stderr, f+"\n", a...)
	}
	os.Exit(1)
}
