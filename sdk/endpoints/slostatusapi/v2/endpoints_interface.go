// Code generated by ifacemaker; DO NOT EDIT.

package v2

import (
	"context"
)

type Endpoints interface {
	GetSLO(ctx context.Context, name, project string) (slo SLODetails, err error)
	GetSLOs(ctx context.Context, limit int, cursor string) (slos SLOListResponse, err error)
}