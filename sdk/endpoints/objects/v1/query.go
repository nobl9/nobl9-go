package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	QueryKeyName              = "name"
	QueryKeyFrom              = "from"
	QueryKeyTo                = "to"
	QueryKeySLOName           = "slo"
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
	f.query.Add(QueryKeyLabels, strings.Join(strLabels, ","))
	return f
}

func (f *filters) Time(key string, t time.Time) *filters {
	if t.IsZero() {
		return f
	}
	f.query.Add(key, t.Format(time.RFC3339))
	return f
}

func (f *filters) Bool(k string, b bool) *filters {
	f.query.Set(k, strconv.FormatBool(b))
	return f
}

func (f *filters) Strings(k string, values []string) *filters {
	for _, v := range values {
		f.query.Add(k, v)
	}
	return f
}

func (f *filters) Floats(k string, values []float64) *filters {
	for _, v := range values {
		f.query.Add(k, strconv.FormatFloat(v, 'f', -1, 64))
	}
	return f
}
