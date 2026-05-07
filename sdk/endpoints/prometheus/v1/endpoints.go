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
	return api.Query(ctx, request.Query, request.Time, queryOptions(request)...)
}

func (e endpoints) QueryRange(ctx context.Context, request QueryRangeRequest) (model.Value, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.QueryRange(
		ctx,
		request.Query,
		promv1.Range{
			Start: request.Start,
			End:   request.End,
			Step:  request.Step,
		},
		queryRangeOptions(request)...)
}

// Source docs: https://prometheus.io/docs/prometheus/latest/querying/api/#getting-label-names
func (e endpoints) LabelNames(ctx context.Context, request LabelNamesRequest) ([]string, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.LabelNames(ctx, request.Matches, request.StartTime, request.EndTime, limitOption(request.Limit)...)
}

func (e endpoints) LabelValues(
	ctx context.Context,
	request LabelValuesRequest,
) (model.LabelValues, promv1.Warnings, error) {
	api, err := e.api(ctx)
	if err != nil {
		return nil, nil, err
	}
	return api.LabelValues(
		ctx,
		request.Label,
		request.Matches,
		request.StartTime,
		request.EndTime,
		limitOption(request.Limit)...)
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

func queryOptions(request QueryRequest) []promv1.Option {
	opts := make([]promv1.Option, 0, 3)
	opts = appendLimitOption(opts, request.Limit)
	opts = appendLookbackDeltaOption(opts, request.LookbackDelta)
	opts = appendTimeoutOption(opts, request.Timeout)
	return opts
}

func queryRangeOptions(request QueryRangeRequest) []promv1.Option {
	opts := make([]promv1.Option, 0, 3)
	opts = appendLimitOption(opts, request.Limit)
	opts = appendLookbackDeltaOption(opts, request.LookbackDelta)
	opts = appendTimeoutOption(opts, request.Timeout)
	return opts
}

func limitOption(limit uint64) []promv1.Option {
	opts := make([]promv1.Option, 0, 1)
	opts = appendLimitOption(opts, limit)
	return opts
}

func appendLimitOption(opts []promv1.Option, limit uint64) []promv1.Option {
	if limit == 0 {
		return opts
	}
	return append(opts, promv1.WithLimit(limit))
}

func appendLookbackDeltaOption(opts []promv1.Option, lookbackDelta time.Duration) []promv1.Option {
	if lookbackDelta == 0 {
		return opts
	}
	return append(opts, promv1.WithLookbackDelta(lookbackDelta))
}

func appendTimeoutOption(opts []promv1.Option, timeout time.Duration) []promv1.Option {
	if timeout == 0 {
		return opts
	}
	return append(opts, promv1.WithTimeout(timeout))
}
