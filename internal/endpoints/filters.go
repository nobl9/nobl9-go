package endpoints

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/nobl9/nobl9-go/internal/sdk"
)

type Filters struct {
	Header http.Header
	Query  url.Values
}

func NewFilters() *Filters {
	return &Filters{
		Header: make(http.Header),
		Query:  make(url.Values),
	}
}

func (f *Filters) Project(project string) *Filters {
	if project == "" {
		return f
	}
	f.Header.Set(sdk.HeaderProject, project)
	return f
}

func (f *Filters) Time(key string, t time.Time) *Filters {
	if t.IsZero() {
		return f
	}
	f.Query.Add(key, t.Format(time.RFC3339))
	return f
}

func (f *Filters) Bool(k string, b *bool) *Filters {
	if b == nil {
		return f
	}
	f.Query.Set(k, strconv.FormatBool(*b))
	return f
}

func (f *Filters) Strings(k string, values []string) *Filters {
	for _, v := range values {
		f.Query.Add(k, v)
	}
	return f
}

func (f *Filters) Floats(k string, values []float64) *Filters {
	for _, v := range values {
		f.Query.Add(k, strconv.FormatFloat(v, 'f', -1, 64))
	}
	return f
}

func (f *Filters) String(k, value string) *Filters {
	if value == "" {
		return f
	}
	f.Query.Set(k, value)
	return f
}
