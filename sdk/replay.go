package sdk

import "github.com/nobl9/nobl9-go/sdk/endpoints/replay"

// Replay is used to access specific Replay API version.
func (c *Client) Replay() replay.Versions {
	return replay.NewVersions(c)
}
