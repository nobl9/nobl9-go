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
	baseAPIPath = "timetravel"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

func (e endpoints) Replay(ctx context.Context, params ReplayRequest) (err error) {
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(params); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	header := http.Header{sdk.HeaderProject: []string{params.Project}}
	req, err := e.client.CreateRequest(ctx, http.MethodPost, baseAPIPath, header, nil, body)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		var httpErr *sdk.HTTPError
		if errors.As(err, &httpErr) && len(httpErr.Errors) == 0 {
			httpErr.Errors[0].Title = replayUnavailabilityReasonExplanation(httpErr.Errors[0].Title)
		}
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
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
