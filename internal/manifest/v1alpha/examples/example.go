package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

type Example interface {
	GetObject() any
	GetVariant() string
	GetSubVariant() string
	GetYAMLComments() []string
}

func newExampleSlice[T Example](tv ...T) []Example {
	examples := make([]Example, 0, len(tv))
	for _, v := range tv {
		examples = append(examples, v)
	}
	return examples
}

type standardExample struct {
	Variant    string
	SubVariant string
	Object     any
}

func (s standardExample) GetObject() any {
	return s.Object
}

func (s standardExample) GetVariant() string {
	return s.Variant
}

func (s standardExample) GetSubVariant() string {
	return s.SubVariant
}

func (s standardExample) GetYAMLComments() []string {
	if s.SubVariant == "" {
		return nil
	}
	return []string{s.SubVariant}
}

type DataSourceTypeGetter interface {
	GetDataSourceType() v1alpha.DataSourceType
}
