package v1

import (
	"context"
	"time"

	"github.com/pkg/errors"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

// APIFactory returns a configured Prometheus-compatible API client.
type APIFactory func(ctx context.Context) (API, error)

// API is the Prometheus-compatible API subset supported by Nobl9.
type API interface {
	Query(ctx context.Context, query string, ts time.Time, opts ...promv1.Option) (model.Value, promv1.Warnings, error)
	QueryRange(
		ctx context.Context,
		query string,
		r promv1.Range,
		opts ...promv1.Option,
	) (model.Value, promv1.Warnings, error)
	Series(
		ctx context.Context,
		matches []string,
		startTime, endTime time.Time,
		opts ...promv1.Option,
	) ([]model.LabelSet, promv1.Warnings, error)
	LabelNames(
		ctx context.Context,
		matches []string,
		startTime, endTime time.Time,
		opts ...promv1.Option,
	) ([]string, promv1.Warnings, error)
	LabelValues(
		ctx context.Context,
		label string,
		matches []string,
		startTime, endTime time.Time,
		opts ...promv1.Option,
	) (model.LabelValues, promv1.Warnings, error)
	Metadata(ctx context.Context, metric, limit string) (map[string][]promv1.Metadata, error)
	Buildinfo(ctx context.Context) (promv1.BuildinfoResult, error)
}

func NewEndpoints(apiFactory APIFactory) Endpoints {
	return endpoints{apiFactory: apiFactory}
}

type endpoints struct {
	apiFactory APIFactory
}

func (e endpoints) Query(ctx context.Context, request QueryRequest) (model.Value, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.Query(ctx, request.Query, request.Time, request.Options...)
}

func (e endpoints) QueryRange(ctx context.Context, request QueryRangeRequest) (model.Value, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.QueryRange(ctx, request.Query, request.Range, request.Options...)
}

func (e endpoints) Series(ctx context.Context, request SeriesRequest) ([]model.LabelSet, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.Series(ctx, request.Matches, request.StartTime, request.EndTime, request.Options...)
}

func (e endpoints) LabelNames(ctx context.Context, request LabelNamesRequest) ([]string, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.LabelNames(ctx, request.Matches, request.StartTime, request.EndTime, request.Options...)
}

func (e endpoints) LabelValues(
	ctx context.Context,
	request LabelValuesRequest,
) (model.LabelValues, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.LabelValues(ctx, request.Label, request.Matches, request.StartTime, request.EndTime, request.Options...)
}

func (e endpoints) Metadata(ctx context.Context, request MetadataRequest) (map[string][]promv1.Metadata, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, err
	}
	return api.Metadata(ctx, request.Metric, request.Limit)
}

func (e endpoints) Buildinfo(ctx context.Context) (promv1.BuildinfoResult, error) {
	api, err := e.api(ctx)
	if err != nil {
		return promv1.BuildinfoResult{}, err
	}
	return api.Buildinfo(ctx)
}

func (e endpoints) api(ctx context.Context) (API, error) {
	if e.apiFactory == nil {
		return nil, errors.New("prometheus api factory is not configured")
	}
	return e.apiFactory(ctx)
}
