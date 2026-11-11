package v1

import "time"

// QueryLogsRequest selects stored, grouped collection logs for one Direct data source.
// From and To bound an inclusive overlap interval of at most 7 days.
// Repeat every filter and sort value unchanged when using PageToken.
type QueryLogsRequest struct {
	Project string    `json:"project"`
	Name    string    `json:"name"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
	// Limit defaults to 50 and must be between 1 and 100 when supplied.
	Limit int `json:"limit,omitempty"`
	// PageToken continues a snapshot using NextPageToken from the latest response.
	// Each continuation renews the snapshot for 5 minutes, up to 30 minutes after the first page.
	// After an invalid or expired token error, restart without a token.
	PageToken string `json:"pageToken,omitempty"`
	// SearchPhrase matches every whitespace-separated word, ignoring case.
	// The wildcard characters * and ? are supported.
	SearchPhrase string              `json:"searchPhrase,omitempty"`
	Levels       []string            `json:"levels,omitempty"`
	Slos         []ResourceReference `json:"slos,omitempty"`
	Sort         LogSort             `json:"sort,omitempty"`
}

// LogSort controls the ordering of grouped logs within a snapshot.
type LogSort struct {
	// Column is firstTimestamp (default), lastTimestamp, level, sloNumber, or message.
	Column string `json:"column,omitempty"`
	// Direction is asc or desc (default).
	Direction string `json:"direction,omitempty"`
}

// ResourceReference identifies a data source or SLO by its project and name.
type ResourceReference struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}

// QueryLogsResponse contains one page, with a serialized response size below 1 MiB.
// A short or empty page can have a NextPageToken when the scan budget was reached.
// Total counts grouped records on this page, not all matches or raw occurrences.
type QueryLogsResponse struct {
	Data          []Log  `json:"data"`
	Total         int    `json:"total"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	// Complete is true only when all matches have been retrieved across visited pages.
	Complete bool `json:"complete"`
	// Incomplete reports a retrieval failure. Restart the query to obtain a new snapshot.
	Incomplete bool `json:"incomplete"`
	// Truncated reports shortened content in at least one returned log.
	Truncated bool `json:"truncated"`
}

// Log is a stored group of repeated messages, not an individual occurrence.
// Timestamps and RepeatedCount describe the entire group, including occurrences
// outside the requested interval. Each serialized log is limited to 8 KiB.
type Log struct {
	Message        string            `json:"message"`
	FirstTimestamp string            `json:"firstTimestamp"`
	LastTimestamp  string            `json:"lastTimestamp"`
	RepeatedCount  uint64            `json:"repeatedCount"`
	Level          string            `json:"level"`
	Datasource     ResourceReference `json:"datasource"`
	Slos           []SLOMetadata     `json:"slos"`
	SloNumber      int               `json:"sloNumber"`
	// Truncated means message or SLO metadata was shortened to fit the response budget.
	Truncated bool `json:"truncated"`
}

// SLOMetadata describes an SLO associated with a grouped log.
type SLOMetadata struct {
	Name        string `json:"name"`
	Project     string `json:"project"`
	DisplayName string `json:"displayName"`
}
