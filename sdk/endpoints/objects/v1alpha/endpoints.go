package v1alpha

import (
	"context"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

type Endpoints struct{}

func (e Endpoints) Apply(ctx context.Context, objects []manifest.Object) error {

}

func (e Endpoints) Delete(ctx context.Context, objects []manifest.Object) error {

}

func (e Endpoints) DeleteProjects(ctx context.Context, names ...string) error {

}

func (e Endpoints) DeleteServices(ctx context.Context, names ...string) error {

}

func (e Endpoints) GetServices(ctx context.Context, params GetServicesRequest) ([]service.Service, error) {

}

func (e Endpoints) GetProjects(ctx context.Context, params GetProjectsRequest) ([]project.Project, error) {

}

func (e Endpoints) GetAlerts(ctx context.Context, params GetAlertsRequest) (*GetAlertsResponse, error) {

}
