//go:build e2e_test

package tests

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

const (
	staleObjectAge      = 3 * time.Hour
	sweepTimeout        = 5 * time.Minute
	sweepDeleteBatch    = 50
	sweepProjectWorkers = 4
)

type sweeper struct {
	now      time.Time
	deleted  atomic.Int64
	projects atomic.Int64
}

// sweepStaleObjects deletes objects leaked by earlier runs which did not reach their cleanup.
// It is best effort: failures are printed and never fail the suite.
func sweepStaleObjects(ctx context.Context, now time.Time) {
	ctx, cancel := context.WithTimeout(ctx, sweepTimeout)
	defer cancel()
	s := &sweeper{now: now}

	s.sweepProjectLess(ctx)

	projects, err := client.Objects().V1().GetV1alphaProjects(ctx, objectsV1.GetProjectsRequest{})
	if err != nil {
		printErrorf("sweep: failed to list Projects: %v", err)
		return
	}
	var staleProjects []string
	for _, project := range projects {
		if name := project.GetName(); s.isStale(name) {
			staleProjects = append(staleProjects, name)
		}
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(sweepProjectWorkers)
	group.Go(func() error {
		s.sweepProjectContents(groupCtx, defaultProject)
		return nil
	})
	for _, name := range staleProjects {
		group.Go(func() error {
			s.sweepProjectContents(groupCtx, name)
			s.deleteInBatches(groupCtx, manifest.KindProject, name, []string{name}, 1)
			return nil
		})
	}
	_ = group.Wait()
	fmt.Printf("Deleted %d stale e2e objects, including %d Projects\n\n", s.deleted.Load(), s.projects.Load())
}

func (s *sweeper) isStale(name string) bool {
	return e2etestutils.IsStaleName(name, s.now, staleObjectAge)
}

func (s *sweeper) sweepProjectLess(ctx context.Context) {
	v1 := client.Objects().V1()
	adjustments, err := v1.GetBudgetAdjustments(ctx, objectsV1.GetBudgetAdjustmentRequest{})
	deleteStale(ctx, s, manifest.KindBudgetAdjustment, "", adjustments, err)
	reports, err := v1.GetReports(ctx, objectsV1.GetReportsRequest{})
	deleteStale(ctx, s, manifest.KindReport, "", reports, err)
	exports, err := v1.GetV1alphaDataExports(ctx, objectsV1.GetDataExportsRequest{Project: sdk.ProjectsWildcard})
	if err != nil {
		printErrorf("sweep: failed to list %s: %v", manifest.KindDataExport, err)
		return
	}
	byProject := make(map[string][]string)
	for _, export := range exports {
		if s.isStale(export.GetName()) {
			byProject[export.GetProject()] = append(byProject[export.GetProject()], export.GetName())
		}
	}
	for project, names := range byProject {
		s.deleteInBatches(ctx, manifest.KindDataExport, project, names, sweepDeleteBatch)
	}
}

// sweepProjectContents deletes stale objects from a Project, in dependency order.
// A Project delete does not cascade and is refused while the Project holds objects.
func (s *sweeper) sweepProjectContents(ctx context.Context, project string) {
	v1 := client.Objects().V1()
	silences, err := v1.GetV1alphaAlertSilences(ctx, objectsV1.GetAlertSilencesRequest{Project: project})
	deleteStale(ctx, s, manifest.KindAlertSilence, project, silences, err)
	slos, err := v1.GetV1alphaSLOs(ctx, objectsV1.GetSLOsRequest{Project: project})
	deleteStale(ctx, s, manifest.KindSLO, project, slos, err)
	policies, err := v1.GetV1alphaAlertPolicies(ctx, objectsV1.GetAlertPolicyRequest{Project: project})
	deleteStale(ctx, s, manifest.KindAlertPolicy, project, policies, err)
	methods, err := v1.GetV1alphaAlertMethods(ctx, objectsV1.GetAlertMethodsRequest{Project: project})
	deleteStale(ctx, s, manifest.KindAlertMethod, project, methods, err)
	services, err := v1.GetV1alphaServices(ctx, objectsV1.GetServicesRequest{Project: project})
	deleteStale(ctx, s, manifest.KindService, project, services, err)
	bindings, err := v1.GetV1alphaRoleBindings(ctx, objectsV1.GetRoleBindingsRequest{Project: project})
	deleteStale(ctx, s, manifest.KindRoleBinding, project, bindings, err)
	agents, err := v1.GetV1alphaAgents(ctx, objectsV1.GetAgentsRequest{Project: project})
	deleteStale(ctx, s, manifest.KindAgent, project, agents, err)
	directs, err := v1.GetV1alphaDirects(ctx, objectsV1.GetDirectsRequest{Project: project})
	deleteStale(ctx, s, manifest.KindDirect, project, directs, err)
	annotations, err := client.Objects().V2().GetV1alphaAnnotations(
		ctx, objectsV2.GetAnnotationsRequest{Project: project})
	deleteStale(ctx, s, manifest.KindAnnotation, project, annotations, err)
}

func deleteStale[T manifest.Object](
	ctx context.Context,
	s *sweeper,
	kind manifest.Kind,
	project string,
	objects []T,
	listErr error,
) {
	if listErr != nil {
		printErrorf("sweep: failed to list %s in Project '%s': %v", kind, project, listErr)
		return
	}
	var names []string
	for _, object := range objects {
		if s.isStale(object.GetName()) {
			names = append(names, object.GetName())
		}
	}
	batch := sweepDeleteBatch
	if kind == manifest.KindAgent || kind == manifest.KindDirect {
		batch = 1
	}
	s.deleteInBatches(ctx, kind, project, names, batch)
}

func (s *sweeper) deleteInBatches(
	ctx context.Context,
	kind manifest.Kind,
	project string,
	names []string,
	batch int,
) {
	for start := 0; start < len(names); start += batch {
		end := min(start+batch, len(names))
		if err := client.Objects().V2().DeleteByName(ctx, objectsV2.DeleteByNameRequest{
			Kind:    kind,
			Project: project,
			Names:   names[start:end],
		}); err != nil {
			printErrorf("sweep: failed to delete %d %s objects in Project '%s': %v", end-start, kind, project, err)
			continue
		}
		s.deleted.Add(int64(end - start))
		if kind == manifest.KindProject {
			s.projects.Add(int64(end - start))
		}
	}
}
