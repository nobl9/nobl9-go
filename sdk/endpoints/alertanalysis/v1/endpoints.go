package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
)

const (
	analysisAPIPath             = "alerting/v1/analysis"
	calculateAlertPolicyAPIPath = "alerting/v1/calculate-alert-policy"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

func (e endpoints) StartAnalysis(
	ctx context.Context,
	params StartAnalysisRequest,
) (response StartAnalysisResponse, err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(params); err != nil {
		return response, errors.Wrap(err, "failed to encode request body")
	}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, analysisAPIPath, nil, nil, buf)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, errors.Wrap(err, "failed to decode response body")
	}
	return response, nil
}

func (e endpoints) CalculateAlertPolicy(
	ctx context.Context,
	params CalculateAlertPolicyRequest,
) (response CalculateAlertPolicyResponse, err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(params); err != nil {
		return response, errors.Wrap(err, "failed to encode request body")
	}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, calculateAlertPolicyAPIPath, nil, nil, buf)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, errors.Wrap(err, "failed to decode response body")
	}
	return response, nil
}

func (e endpoints) GetAnalysis(
	ctx context.Context,
	params GetAnalysisRequest,
) (response GetAnalysisResponse, err error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		path.Join(analysisAPIPath, params.AnalysisID),
		nil,
		params.queryValues(),
		nil,
	)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, errors.Wrap(err, "failed to decode response body")
	}
	return response, nil
}

func (e endpoints) RetryAnalysis(
	ctx context.Context,
	analysisID string,
) (response StartAnalysisResponse, err error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPost,
		path.Join(analysisAPIPath, analysisID, "retry"),
		nil,
		nil,
		nil,
	)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, errors.Wrap(err, "failed to decode response body")
	}
	return response, nil
}

func (r GetAnalysisRequest) queryValues() url.Values {
	q := url.Values{}
	if r.From != nil {
		q.Set("from", r.From.Format(time.RFC3339Nano))
	}
	if r.To != nil {
		q.Set("to", r.To.Format(time.RFC3339Nano))
	}
	if r.IncludeTimeseries != nil {
		q.Set("includeTimeseries", strconv.FormatBool(*r.IncludeTimeseries))
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
