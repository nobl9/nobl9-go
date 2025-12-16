package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/sdk"
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

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

func (e endpoints) Run(ctx context.Context, params RunRequest) (err error) {
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(params); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	header := http.Header{sdk.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, apiCreateReplay, header, nil, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return handleReplayError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func (e endpoints) Delete(ctx context.Context, params DeleteRequest) (err error) {
	return e.deleteReplay(ctx, internalDeleteRequest{DeleteRequest: params})
}

func (e endpoints) DeleteAll(ctx context.Context) (err error) {
	return e.deleteReplay(ctx, internalDeleteRequest{All: true})
}

func (e endpoints) Cancel(ctx context.Context, params DeleteRequest) (err error) {
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(params); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	header := http.Header{sdk.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, apiCancelReplay, header, nil, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return handleReplayError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func (e endpoints) List(ctx context.Context) ([]ReplayListItem, error) {
	req, err := e.client.CreateRequest(ctx, http.MethodGet, apiListReplays, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, handleReplayError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var list []ReplayListItem
	if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, errors.Wrap(err, "cannot decode list Replays response")
	}
	return list, nil
}

func (e endpoints) GetStatus(ctx context.Context, params GetStatusRequest) (*ReplayWithStatus, error) {
	path := fmt.Sprintf(apiReplayStatus, params.SLO)
	header := http.Header{sdk.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodGet, path, header, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, handleReplayError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var list []ReplayListItem
	if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, errors.Wrap(err, "cannot decode list Replays response")
	}
	return list, nil
}

func (e endpoints) deleteReplay(ctx context.Context, params internalDeleteRequest) (err error) {
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(params); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	header := http.Header{sdk.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodDelete, apiDeleteReplay, header, nil, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return handleReplayError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func handleReplayError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *sdk.HTTPError
	if errors.As(err, &httpErr) && len(httpErr.Errors) == 0 {
		httpErr.Errors[0].Title = replayUnavailabilityReasonExplanation(httpErr.Errors[0].Title)
	}
	return err
}

func replayUnavailabilityReasonExplanation(reason string) string {
	switch reason {
	case ReplayIntegrationDoesNotSupportReplay:
		return "The Data Source does not support Replay yet"
	case ReplayAgentVersionDoesNotSupportReplay:
		return "Update your Agent version to the latest to use Replay for this Data Source."
	case ReplayMaxHistoricalDataRetrievalTooLow:
		return "Value configured for spec.historicalDataRetrieval.maxDuration.value" +
			" for the Data Source is lower than the duration you're trying to run Replay for."
	case ReplayConcurrentReplayRunsLimitExhausted:
		return "You've exceeded the limit of concurrent Replay runs. Wait until the current Replay(s) are done."
	case ReplayUnknownAgentVersion:
		return "Your Agent isn't connected to the Data Source. Deploy the Agent and run Replay once again."
	case "single_query_not_supported":
		return "Historical data retrieval for single-query ratio metrics is not supported"
	case "composite_slo_not_supported":
		return "Historical data retrieval for Composite SLO is not supported"
	case "promql_in_gcm_not_supported":
		return "Historical data retrieval for PromQL metrics is not supported"
	default:
		return reason
	}
}
