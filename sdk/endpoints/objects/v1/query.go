package v1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	QueryKeyName              = "name"
	QueryKeyFrom              = "from"
	QueryKeyTo                = "to"
	QueryKeySLOName           = "slo"
	QueryKeySLOProjectName    = "slo_project"
	QueryKeyLabels            = "labels"
	QueryKeyServiceName       = "service_name"
	QueryKeyDryRun            = "dry_run"
	QueryKeyAlertPolicyName   = "alert_policy"
	QueryKeyObjectiveName     = "objective"
	QueryKeyObjectiveValue    = "objective_value"
	QueryKeyResolved          = "resolved"
	QueryKeyTriggered         = "triggered"
	QueryKeySystemAnnotations = "system_annotations"
	QueryKeyUserAnnotations   = "user_annotations"
	QueryKeyCategory          = "category"
	queryKeyPaginationLimit   = "pagination.limit"
	queryKeyPaginationOffset  = "pagination.offset"
	queryKeySortColumn        = "sort.column"
	queryKeySortDirection     = "sort.direction"
)

type filters struct {
	*endpointsHelpers.Filters
}

func filterBy() *filters {
	return &filters{Filters: endpointsHelpers.NewFilters()}
}

func (f *filters) Project(project string) *filters {
	f.Filters.Project(project)
	return f
}

func (f *filters) Labels(labels v1alpha.Labels) *filters {
	if labels == nil {
		return f
	}
	var strLabels []string
	for key, values := range labels {
		if len(values) > 0 {
			for _, value := range values {
				strLabels = append(strLabels, fmt.Sprintf("%s:%s", key, value))
			}
		} else {
			strLabels = append(strLabels, key)
		}
	}
	f.Query.Add(QueryKeyLabels, strings.Join(strLabels, ","))
	return f
}

func (f *filters) Time(key string, t time.Time) *filters {
	f.Filters.Time(key, t)
	return f
}

func (f *filters) Bool(k string, b *bool) *filters {
	f.Filters.Bool(k, b)
	return f
}

func (f *filters) Strings(k string, values []string) *filters {
	f.Filters.Strings(k, values)
	return f
}

func (f *filters) Floats(k string, values []float64) *filters {
	f.Filters.Floats(k, values)
	return f
}

func (f *filters) String(k, value string) *filters {
	f.Filters.String(k, value)
	return f
}

func (r GetSLOsRequest) addListQuery(f *filters) {
	if r.Pagination != nil {
		if r.Pagination.Limit > 0 {
			f.String(queryKeyPaginationLimit, strconv.Itoa(r.Pagination.Limit))
		}
		if r.Pagination.Offset > 0 {
			f.String(queryKeyPaginationOffset, strconv.Itoa(r.Pagination.Offset))
		}
	}
	if r.Sort != nil {
		f.String(queryKeySortColumn, string(r.Sort.Column))
		f.String(queryKeySortDirection, string(r.Sort.Direction))
	}
}
