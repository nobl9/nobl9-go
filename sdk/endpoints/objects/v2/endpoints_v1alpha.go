package v2

import (
	"context"
	"encoding/json"
	"net/http"

	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
)

func (e endpoints) GetV1alphaAnnotations(
	ctx context.Context,
	params GetAnnotationsRequest,
) ([]v1alphaAnnotation.Annotation, error) {
	var categories []string
	if len(params.Categories) > 0 {
		categories = make([]string, 0, len(params.Categories))
		for _, category := range params.Categories {
			categories = append(categories, category.String())
		}
	}
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names).
		Time(QueryKeyFrom, params.From).
		Time(QueryKeyTo, params.To).
		Strings(QueryKeyCategory, categories).
		String(QueryKeySLOName, params.SLOName)
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		apiGetAnnotations,
		f.Header,
		f.Query,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var annotations []getAnnotationModel
	if err := json.NewDecoder(resp.Body).Decode(&annotations); err != nil {
		return nil, err
	}
	v1alphaAnnotations := make([]v1alphaAnnotation.Annotation, 0, len(annotations))
	for _, annotation := range annotations {
		v1alphaAnnotations = append(v1alphaAnnotations, getAnnotationsModelToV1alpha(annotation))
	}
	return v1alphaAnnotations, nil
}
