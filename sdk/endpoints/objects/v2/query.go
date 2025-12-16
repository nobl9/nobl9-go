package v2

import (
	"time"

	internalendpoints "github.com/nobl9/nobl9-go/internal/endpoints"
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
	*internalendpoints.Filters
}

func filterBy() *filters {
	return &filters{Filters: internalendpoints.NewFilters()}
}

func (f *filters) Project(project string) *filters {
	f.Filters.Project(project)
	return f
}

func (f *filters) Time(key string, t time.Time) *filters {
	f.Filters.Time(key, t)
	return f
}

func (f *filters) Strings(k string, values []string) *filters {
	f.Filters.Strings(k, values)
	return f
}

func (f *filters) String(k, value string) *filters {
	f.Filters.String(k, value)
	return f
}
