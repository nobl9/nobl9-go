//go:build e2e_test

package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	datasourceV1 "github.com/nobl9/nobl9-go/sdk/endpoints/datasource/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_DataSource_V1_QueryLogs(t *testing.T) {
	now := time.Now().UTC()
	query := datasourceV1.QueryLogsRequest{
		Project: defaultProject, Name: e2etestutils.GenerateName(),
		From: now.Add(-time.Hour), To: now, Limit: 2,
	}
	response, err := client.DataSource().V1().QueryLogs(t.Context(), query)
	require.NoError(t, err)
	require.Empty(t, response.Data)
	require.Zero(t, response.Total)
	require.True(t, response.Complete)
	require.False(t, response.Incomplete)
	require.Empty(t, response.NextPageToken)
	query.PageToken = "invalid"
	_, err = client.DataSource().V1().QueryLogs(t.Context(), query)
	require.Error(t, err)
}
