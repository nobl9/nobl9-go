//go:build e2e_test

package tests

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SLOStatusAPI_V1_GetSLO(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	response, err := client.SLOStatusAPI().V1().GetSLO(ctx, "", "") // TODO PC-14087: Add name and project here.
	require.NoError(t, err)

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal response, err: %v", err)
	}

	assert.NotEmpty(t, string(jsonData))
}

func Test_SLOStatusAPI_V1_GetSLOList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	response, err := client.SLOStatusAPI().V1().GetSLOList(ctx, 2.0, "")
	require.NoError(t, err)

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal response, err: %v", err)
	}

	assert.NotEmpty(t, string(jsonData))
}

func Test_SLOStatusAPI_V1_GetSLOList_Pagination(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	initialResponse, err := client.SLOStatusAPI().V1().GetSLOList(ctx, 2.0, "")
	require.NoError(t, err)

	initialData, err := json.MarshalIndent(initialResponse, "", "  ")
	require.NoError(t, err)

	assert.NotEmpty(t, string(initialData))

	// Extract the cursor from the initial response.
	var initialResult struct {
		Links struct {
			Cursor string `json:"cursor"`
		} `json:"links"`
	}
	err = json.Unmarshal(initialData, &initialResult)
	require.NoError(t, err)
	cursor := initialResult.Links.Cursor
	require.NotEmpty(t, cursor)

	// Make the second request using the cursor to get the next batch of SLOs.
	nextResponse, err := client.SLOStatusAPI().V1().GetSLOList(ctx, 2.0, cursor)
	require.NoError(t, err)

	nextData, err := json.MarshalIndent(nextResponse, "", "  ")
	require.NoError(t, err)

	assert.NotEmpty(t, string(nextData))

	// Verify that the second batch of SLOs is different from the first batch.
	assert.NotEqual(t, string(initialData), string(nextData))
}
