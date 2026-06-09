package v1

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalSDK "github.com/nobl9/nobl9-go/internal/sdk"
)

func TestEndpoints_GetStatus(t *testing.T) {
	t.Parallel()

	client := &replayClientStub{
		responseBody: `{
			"project": "project-a",
			"slo": "slo-a",
			"status": {
				"status": "completed",
				"unit": "Minute",
				"value": 15
			}
		}`,
	}
	endpoints := NewEndpoints(client)

	status, err := endpoints.GetStatus(context.Background(), GetStatusRequest{
		Project: "project-a",
		SLO:     "slo-a",
	})

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, http.MethodGet, client.request.Method)
	assert.Equal(t, "/timetravel/slo-a", client.request.URL.Path)
	assert.Equal(t, "project-a", client.request.Header.Get(internalSDK.HeaderProject))
	assert.Equal(t, "project-a", status.Project)
	assert.Equal(t, "slo-a", status.SLO)
	assert.Equal(t, ReplayStatusCompleted, status.Status.Status)
	assert.Equal(t, "Minute", status.Status.Unit)
	assert.Equal(t, 15, status.Status.Value)
}

func TestEndpoints_GetAvailability(t *testing.T) {
	t.Parallel()

	client := &replayClientStub{
		responseBody: `{"available": false, "reason": "integration_does_not_support_replay"}`,
	}
	endpoints := NewEndpoints(client)

	availability, err := endpoints.GetAvailability(context.Background(), GetAvailabilityRequest{
		Project:           "project-a",
		DataSourceProject: "data-source-project",
		DataSource:        "prometheus",
		DataSourceKind:    "Agent",
		SLOName:           "slo-a",
		Type:              "recalculation",
		DurationUnit:      DurationUnitHour,
		DurationValue:     6,
	})

	require.NoError(t, err)
	require.NotNil(t, availability)
	assert.Equal(t, http.MethodGet, client.request.Method)
	assert.Equal(t, "/internal/timemachine/availability", client.request.URL.Path)
	assert.Equal(t, "project-a", client.request.Header.Get(internalSDK.HeaderProject))
	assert.Equal(t, "data-source-project", client.request.URL.Query().Get("dataSourceProject"))
	assert.Equal(t, "prometheus", client.request.URL.Query().Get("dataSource"))
	assert.Equal(t, "Agent", client.request.URL.Query().Get("dataSourceKind"))
	assert.Equal(t, "slo-a", client.request.URL.Query().Get("sloName"))
	assert.Equal(t, "recalculation", client.request.URL.Query().Get("type"))
	assert.Equal(t, DurationUnitHour, client.request.URL.Query().Get("durationUnit"))
	assert.Equal(t, "6", client.request.URL.Query().Get("durationValue"))
	assert.False(t, availability.Available)
	assert.Equal(t, ReplayIntegrationDoesNotSupportReplay, availability.Reason)
}

type replayClientStub struct {
	request      *http.Request
	responseBody string
	err          error
}

func (c *replayClientStub) CreateRequest(
	ctx context.Context,
	method, endpoint string,
	headers http.Header,
	q url.Values,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		"https://api.nobl9.test/"+endpoint,
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func (c *replayClientStub) Do(req *http.Request) (*http.Response, error) {
	c.request = req
	if c.err != nil {
		return nil, c.err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(c.responseBody)),
	}, nil
}
