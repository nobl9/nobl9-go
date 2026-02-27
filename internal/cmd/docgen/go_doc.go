package main

import (
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"io/fs"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/nobl9/nobl9-go/internal/pathutils"
)

const moduleRootPath = "github.com/nobl9/nobl9-go"

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
	root := pathutils.FindModuleRoot()
	objectsDirectory := filepath.Join(root, "manifest")
	directories, err := listDirectories(objectsDirectory)
	if err != nil {
		log.Panicf("Error listing directories under %s: %v", objectsDirectory, err)
	}

	typeDocs := make(map[string]goTypeDoc, 500)
	for _, dir := range directories {
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
			Dir:  dir,
		}
		pkgs, err := packages.Load(cfg, ".")
		if err != nil {
			log.Panicf("Error loading package in directory %s: %v", dir, err)
		}
		if len(pkgs) == 0 {
			continue
		}
		if len(pkgs) > 1 {
			log.Panicf("Expected exactly one package in %s, got %d", dir, len(pkgs))
		}
		pkg := pkgs[0]
		// Skip directories without Go files or with loading errors
		if len(pkg.Errors) > 0 || len(pkg.Syntax) == 0 {
			continue
		}

		// Convert packages.Package to ast.Package
		// nolint: staticcheck
		astPkg := &ast.Package{
			Name:  pkg.Name,
			Files: make(map[string]*ast.File),
		}
		for i, file := range pkg.Syntax {
			if i < len(pkg.GoFiles) && parserFilter(pkg.GoFiles[i]) {
				astPkg.Files[pkg.GoFiles[i]] = file
			}
		}

		if len(astPkg.Files) == 0 {
			continue
		}

		pkgParser := newPackageParser(astPkg, strings.TrimPrefix(dir, root+"/"))
		docs := pkgParser.Parse()
		for k, v := range docs {
			typeDocs[k] = v
		}
	}
	return typeDocs
}

// nolint: staticcheck
func newPackageParser(astPkg *ast.Package, relPath string) packageParser {
	pkgDoc := doc.New(astPkg, ".", doc.AllDecls)
	return packageParser{
		Path:          filepath.Join(moduleRootPath, relPath),
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
		if selector, ok := fieldType.(*ast.SelectorExpr); ok {
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

func parserFilter(filename string) bool {
	name := filepath.Base(filename)
	return !strings.HasSuffix(name, "_test.go")
}
