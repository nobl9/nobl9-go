package main

import (
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/nobl9/nobl9-go/manifest"

	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlert "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1alphaUserGroup "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
)

const objectsDirectory = "manifest"

func main() {
	objects := []manifest.Object{
		v1alphaProject.Project{},
		v1alphaService.Service{},
		v1alphaSLO.SLO{},
		v1alphaDirect.Direct{},
		v1alphaAgent.Agent{},
		v1alphaAlertMethod.AlertMethod{},
		v1alphaAlertPolicy.AlertPolicy{},
		v1alphaAlertSilence.AlertSilence{},
		v1alphaAlert.Alert{},
		v1alphaAnnotation.Annotation{},
		v1alphaBudgetAdjustment.BudgetAdjustment{},
		v1alphaDataExport.DataExport{},
		v1alphaUserGroup.UserGroup{},
		v1alphaRoleBinding.RoleBinding{},
	}
	objects = objects

	directories, err := listDirectories(objectsDirectory)
	if err != nil {
		log.Panicf("Error listing directories under %s: %v", objectsDirectory, err)
	}

	for _, dir := range directories {
		fset := token.NewFileSet()
		packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
		if err != nil {
			log.Panicf("Error parsing directory %s: %v", dir, err)
			return
		}

		for name, pkg := range packages {
			fmt.Printf("Parsing package %s\n", name)
			pkgDoc := doc.New(pkg, ".", doc.AllDecls)
			for _, t := range pkgDoc.Types {
				fmt.Println("Struct documentation:")
				fmt.Println(t.Doc)
			}
			fmt.Println("Struct not found")
		}
	}
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
