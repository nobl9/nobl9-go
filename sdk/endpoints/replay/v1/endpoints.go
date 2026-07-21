package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	internalSDK "github.com/nobl9/nobl9-go/internal/sdk"
)

const (
	apiCreateReplay       = "timetravel"
	apiDeleteReplay       = "timetravel"
	apiCancelReplay       = "timetravel/cancel"
	apiListReplays        = "timetravel/list"
	apiReplayStatus       = "timetravel/%s"
	apiReplayAvailability = "internal/timemachine/availability"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

// NewEndpoints creates replay V1 endpoints backed by client.
func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

// Run validates params and starts a replay.
func (e endpoints) Run(ctx context.Context, params RunRequest) (err error) {
	if err = params.Validate(); err != nil {
		return err
	}
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(params); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	header := http.Header{internalSDK.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, apiCreateReplay, header, nil, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// Delete removes queued replay requests selected by params.
func (e endpoints) Delete(ctx context.Context, params DeleteRequest) (err error) {
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(params); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	header := http.Header{internalSDK.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodDelete, apiDeleteReplay, header, nil, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// Cancel requests cancellation of the replay selected by params.
func (e endpoints) Cancel(ctx context.Context, params CancelRequest) (err error) {
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(params); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	header := http.Header{internalSDK.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, apiCancelReplay, header, nil, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// List returns queued and in-progress reimport-and-recalculation replays visible
// to the current organization.
func (e endpoints) List(ctx context.Context) ([]ReplayListItem, error) {
	req, err := e.client.CreateRequest(ctx, http.MethodGet, apiListReplays, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var list []ReplayListItem
	if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, errors.Wrap(err, "cannot decode list Replays response")
	}
	return list, nil
}

// GetStatus returns the latest detailed replay status for an SLO.
func (e endpoints) GetStatus(ctx context.Context, params GetStatusRequest) (*ReplayWithStatus, error) {
	path := fmt.Sprintf(apiReplayStatus, params.SLO)
	header := http.Header{internalSDK.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodGet, path, header, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var status ReplayWithStatus
	if err = json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, errors.Wrap(err, "cannot decode Replay status response")
	}
	return &status, nil
}

// GetAvailability validates params and reports whether a replay can be started.
func (e endpoints) GetAvailability(
	ctx context.Context,
	params GetAvailabilityRequest,
) (*ReplayAvailability, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	header := http.Header{internalSDK.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		apiReplayAvailability,
		header,
		params.queryValues(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var availability ReplayAvailability
	if err = json.NewDecoder(resp.Body).Decode(&availability); err != nil {
		return nil, errors.Wrap(err, "cannot decode Replay availability response")
	}
	return &availability, nil
}

func (r GetAvailabilityRequest) queryValues() url.Values {
	q := url.Values{}
	if r.DataSourceProject != "" {
		q.Set("dataSourceProject", r.DataSourceProject)
	}
	if r.DataSource != "" {
		q.Set("dataSource", r.DataSource)
	}
	if r.DataSourceKind != "" {
		q.Set("dataSourceKind", r.DataSourceKind)
	}
	if r.SLOName != "" {
		q.Set("sloName", r.SLOName)
	}
	if r.Type != "" {
		q.Set("type", string(r.Type))
	}
	if r.DurationUnit != "" {
		q.Set("durationUnit", string(r.DurationUnit))
	}
	if r.DurationValue != 0 {
		q.Set("durationValue", strconv.Itoa(r.DurationValue))
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
