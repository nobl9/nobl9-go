package v2

import (
	"net/http"
	"net/url"
	"time"

	"github.com/nobl9/nobl9-go/internal/sdk"
)

const (
	QueryKeyName     = "name"
	QueryKeyFrom     = "from"
	QueryKeyTo       = "to"
	QueryKeySLOName  = "slo"
	QueryKeyCategory = "category"
	QueryKeyDryRun   = "dry_run"
)

type filters struct {
	header http.Header
	query  url.Values
}

func filterBy() *filters {
	return &filters{
		header: make(http.Header),
		query:  make(url.Values),
	}
}

func (f *filters) Project(project string) *filters {
	if project == "" {
		return f
	}
	f.header.Set(sdk.HeaderProject, project)
	return f
}

func (f *filters) Time(key string, t time.Time) *filters {
	if t.IsZero() {
		return f
	}
	f.query.Add(key, t.Format(time.RFC3339))
	return f
}

func (f *filters) Strings(k string, values []string) *filters {
	for _, v := range values {
		f.query.Add(k, v)
	}
	return f
}

func (f *filters) String(k, value string) *filters {
	if value == "" {
		return f
	}
	f.query.Set(k, value)
	return f
}
