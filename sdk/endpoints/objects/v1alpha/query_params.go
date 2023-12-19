package v1alpha

import (
	"fmt"
	"github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	QueryKeyName              = "name"
	QueryKeyFrom              = "from"
	QueryKeyTo                = "to"
	QueryKeySLOName           = "slo"
	QueryKeyLabelsFilter      = "labels"
	QueryKeyServiceName       = "service_name"
	QueryKeyDryRun            = "dry_run"
	QueryKeyAlertPolicyName   = "alert_policy"
	QueryKeyObjectiveName     = "objective"
	QueryKeyObjectiveValue    = "objective_value"
	QueryKeyResolved          = "resolved"
	QueryKeyTriggered         = "triggered"
	QueryKeySystemAnnotations = "system_annotations"
	QueryKeyUserAnnotations   = "user_annotations"
)

type Filters struct {
	header http.Header
	query  url.Values
}

func FilterBy() *Filters {
	return &Filters{
		header: make(http.Header),
		query:  make(url.Values),
	}
}

func (f *Filters) Project(project string) *Filters {
	f.header.Set(sdk.HeaderProject, project)
	return f
}

func (f *Filters) DryRun() {
	f.query.Set(QueryKeyDryRun, strconv.FormatBool(true))
}

func (f *Filters) Names(names ...string) *Filters {
	for _, name := range names {
		f.query.Add(QueryKeyName, name)
	}
	return f
}

func (f *Filters) Labels(labels v1alpha.Labels) *Filters {
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
	f.query.Add(QueryKeyLabelsFilter, strings.Join(strLabels, ","))
	return f
}

func (f *Filters) From(from time.Time) *Filters {
	f.query.Add(QueryKeyFrom, from.Format(time.RFC3339))
	return f
}

func (f *Filters) To(to time.Time) *Filters {
	f.query.Add(QueryKeyTo, to.Format(time.RFC3339))
	return f
}

func (f *Filters) SLONames(names ...string) *Filters {
	for _, name := range names {
		f.query.Add(QueryKeySLOName, name)
	}
	return f
}

func (f *Filters) ServiceNames(names ...string) *Filters {
	for _, name := range names {
		f.query.Add(QueryKeyServiceName, name)
	}
	return f
}

func (f *Filters) AlertPolicyNames(names ...string) *Filters {
	for _, name := range names {
		f.query.Add(QueryKeyAlertPolicyName, name)
	}
	return f
}

func (f *Filters) AlertObjectiveNames(names ...string) *Filters {
	for _, name := range names {
		f.query.Add(QueryKeyObjectiveName, name)
	}
	return f
}

func (f *Filters) AlertObjectiveValues(values ...string) *Filters {
	for _, value := range values {
		f.query.Add(QueryKeyObjectiveValue, value)
	}
	return f
}

func (f *Filters) ResolvedAlerts() *Filters {
	f.query.Add(QueryKeyResolved, "true")
	return f
}

func (f *Filters) TriggeredAlerts() *Filters {
	f.query.Add(QueryKeyTriggered, "true")
	return f
}
