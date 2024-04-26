package main

import (
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/exp/maps"
)

const (
	objectsDirectory = "manifest"
	moduleRootPath   = "github.com/nobl9/nobl9-go"
)

type goTypeDoc struct {
	Name         string
	Package      string
	Doc          string
	StructFields map[string]goTypeDoc
}

func (t goTypeDoc) PkgPath() string {
	return filepath.Join(t.Package, t.Name)
}

func parseGoDocs() map[string]goTypeDoc {
	directories, err := listDirectories(objectsDirectory)
	if err != nil {
		log.Panicf("Error listing directories under %s: %v", objectsDirectory, err)
	}

	typeDocs := make(map[string]goTypeDoc, 500)
	for _, dir := range directories {
		packages, err := parser.ParseDir(token.NewFileSet(), dir, parserFilter, parser.ParseComments)
		if err != nil {
			log.Panicf("Error parsing directory %s: %v", dir, err)
		}
		if len(packages) == 0 {
			continue
		}
		if len(packages) > 1 {
			log.Panicf("Expected exactly one package in %s, got %d", dir, len(packages))
		}
		pkg := maps.Values(packages)[0]
		pkgParser := newPackageParser(pkg, dir)
		docs := pkgParser.Parse()
		for k, v := range docs {
			typeDocs[k] = v
		}
	}
	return typeDocs
}

func newPackageParser(astPkg *ast.Package, dir string) packageParser {
	pkgDoc := doc.New(astPkg, ".", doc.AllDecls)
	return packageParser{
		Path:          filepath.Join(moduleRootPath, dir),
		Doc:           pkgDoc,
		CommentParser: pkgDoc.Parser(),
	}
}

type packageParser struct {
	Path          string
	Doc           *doc.Package
	CommentParser *comment.Parser
}

var startsWithLowercaseRegexp = regexp.MustCompile("^[a-z]")

func (p packageParser) Parse() map[string]goTypeDoc {
	typeDocs := make(map[string]goTypeDoc, len(p.Doc.Types))
	for _, typ := range p.Doc.Types {
		// Skip unexported types.
		if startsWithLowercaseRegexp.MatchString(typ.Name) {
			continue
		}
		td := goTypeDoc{
			Name:    typ.Name,
			Package: p.Path,
			Doc:     typ.Doc,
		}
		if len(typ.Decl.Specs) > 0 {
			td.StructFields = p.parseStructFields(typ.Decl.Specs[0])
		}
		if td.Doc == "" && len(td.StructFields) == 0 {
			continue
		}
		typeDocs[td.PkgPath()] = td
	}
	return typeDocs
}

func (p packageParser) parseStructFields(spec ast.Spec) map[string]goTypeDoc {
	typSpec, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil
	}
	structType, ok := typSpec.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	fields := make(map[string]goTypeDoc, len(structType.Fields.List))
	for _, field := range structType.Fields.List {
		if field.Doc == nil || field.Tag == nil {
			continue
		}
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json")
		tagValues := strings.Split(tag, ",")
		if len(tagValues) == 0 {
			continue
		}
		tagName := tagValues[0]
		if tagName == "" || tagName == "-" {
			continue
		}
		ftd := goTypeDoc{
			Name:    field.Names[0].Name,
			Package: p.Path,
			Doc:     field.Doc.Text(),
		}
		fieldType := field.Type
		// Extract pointer type if present.
		if star, ok := fieldType.(*ast.StarExpr); ok {
			fieldType = star.X
		}
		if selector, ok := field.Type.(*ast.SelectorExpr); ok {
			fieldPkgName := selector.X.(*ast.Ident).Name
			importPath, found := p.CommentParser.LookupPackage(fieldPkgName)
			if !found {
				log.Panicf("Package not found: %s", fieldPkgName)
			}
			ftd.Package = importPath
		}
		fields[tagName] = ftd
	}
	return fields
}

func listDirectories(root string) ([]string, error) {
	directories := make([]string, 0, 20)
	err := filepath.WalkDir(root, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() {
			directories = append(directories, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return directories, nil
}

func parserFilter(info fs.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}
