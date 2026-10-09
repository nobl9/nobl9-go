//go:build e2e_test

package tests

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

const (
	e2eNamePrefix       = "sdk-e2e-"
	staleObjectAge      = 3 * time.Hour
	sweepTimeout        = 5 * time.Minute
	sweepDeleteBatch    = 50
	sweepProjectWorkers = 4
)

// sweepStaleObjects deletes objects leaked by earlier runs which did not reach their cleanup.
// It is best effort: failures are printed and never fail the suite.
func sweepStaleObjects(ctx context.Context, now time.Time) {
	ctx, cancel := context.WithTimeout(ctx, sweepTimeout)
	defer cancel()

	var swept int
	exports, err := client.Objects().V1().GetV1alphaDataExports(
		ctx, objectsV1.GetDataExportsRequest{Project: sdk.ProjectsWildcard})
	if err != nil {
		printErrorf("sweep: failed to list Data Exports: %v", err)
	}
	exportsByProject := make(map[string][]string)
	for _, export := range exports {
		if isStaleE2EName(export.GetName(), now) {
			project := export.GetProject()
			exportsByProject[project] = append(exportsByProject[project], export.GetName())
		}
	}
	for project, names := range exportsByProject {
		swept += len(names)
		deleteInBatches(ctx, manifest.KindDataExport, project, names, sweepDeleteBatch)
	}

	projects, err := client.Objects().V1().GetV1alphaProjects(ctx, objectsV1.GetProjectsRequest{})
	if err != nil {
		printErrorf("sweep: failed to list Projects: %v", err)
		return
	}
	var stale []string
	for _, project := range projects {
		name := project.GetName()
		if name == defaultProject || name == e2etestutils.DataSourcesProject {
			continue
		}
		if isStaleE2EName(name, now) {
			stale = append(stale, name)
		}
	}
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(sweepProjectWorkers)
	for _, name := range stale {
		group.Go(func() error {
			sweepProject(groupCtx, name)
			return nil
		})
	}
	_ = group.Wait()
	fmt.Printf("Swept stale e2e objects: %d Data Exports, %d Projects\n\n", swept, len(stale))
}

func sweepProject(ctx context.Context, project string) {
	v1 := client.Objects().V1()
	steps := []struct {
		kind  manifest.Kind
		batch int
		list  func() ([]string, error)
	}{
		{manifest.KindAlertSilence, sweepDeleteBatch, func() ([]string, error) {
			found, err := v1.GetV1alphaAlertSilences(ctx, objectsV1.GetAlertSilencesRequest{Project: project})
			return objectNames(found), err
		}},
		{manifest.KindSLO, sweepDeleteBatch, func() ([]string, error) {
			found, err := v1.GetV1alphaSLOs(ctx, objectsV1.GetSLOsRequest{Project: project})
			return objectNames(found), err
		}},
		{manifest.KindAlertPolicy, sweepDeleteBatch, func() ([]string, error) {
			found, err := v1.GetV1alphaAlertPolicies(ctx, objectsV1.GetAlertPolicyRequest{Project: project})
			return objectNames(found), err
		}},
		{manifest.KindAlertMethod, sweepDeleteBatch, func() ([]string, error) {
			found, err := v1.GetV1alphaAlertMethods(ctx, objectsV1.GetAlertMethodsRequest{Project: project})
			return objectNames(found), err
		}},
		{manifest.KindService, sweepDeleteBatch, func() ([]string, error) {
			found, err := v1.GetV1alphaServices(ctx, objectsV1.GetServicesRequest{Project: project})
			return objectNames(found), err
		}},
		{manifest.KindAgent, 1, func() ([]string, error) {
			found, err := v1.GetV1alphaAgents(ctx, objectsV1.GetAgentsRequest{Project: project})
			return objectNames(found), err
		}},
		{manifest.KindDirect, 1, func() ([]string, error) {
			found, err := v1.GetV1alphaDirects(ctx, objectsV1.GetDirectsRequest{Project: project})
			return objectNames(found), err
		}},
		{manifest.KindAnnotation, sweepDeleteBatch, func() ([]string, error) {
			found, err := client.Objects().V2().GetV1alphaAnnotations(
				ctx, objectsV2.GetAnnotationsRequest{Project: project})
			return objectNames(found), err
		}},
	}
	for _, step := range steps {
		names, err := step.list()
		if err != nil {
			printErrorf("sweep: failed to list %s in Project %s: %v", step.kind, project, err)
			continue
		}
		deleteInBatches(ctx, step.kind, project, names, step.batch)
	}
	if err := client.Objects().V2().DeleteByName(ctx, objectsV2.DeleteByNameRequest{
		Kind:    manifest.KindProject,
		Project: project,
		Names:   []string{project},
	}); err != nil {
		printErrorf("sweep: failed to delete Project %s: %v", project, err)
	}
}

func deleteInBatches(ctx context.Context, kind manifest.Kind, project string, names []string, batch int) {
	for start := 0; start < len(names); start += batch {
		end := min(start+batch, len(names))
		if err := client.Objects().V2().DeleteByName(ctx, objectsV2.DeleteByNameRequest{
			Kind:    kind,
			Project: project,
			Names:   names[start:end],
		}); err != nil {
			printErrorf("sweep: failed to delete %d %s objects in Project %s: %v", end-start, kind, project, err)
		}
	}
}

func objectNames[T manifest.Object](objects []T) []string {
	names := make([]string, 0, len(objects))
	for _, object := range objects {
		names = append(names, object.GetName())
	}
	return names
}

// isStaleE2EName reports whether name was generated by [e2etestutils.GenerateName]
// more than [staleObjectAge] before now. The generated name ends with its creation time in Unix nanoseconds.
func isStaleE2EName(name string, now time.Time) bool {
	if !strings.HasPrefix(name, e2eNamePrefix) {
		return false
	}
	idx := strings.LastIndexByte(name, '-')
	nanos, err := strconv.ParseInt(name[idx+1:], 10, 64)
	if err != nil || nanos <= 0 {
		return false
	}
	return now.Sub(time.Unix(0, nanos)) > staleObjectAge
}
