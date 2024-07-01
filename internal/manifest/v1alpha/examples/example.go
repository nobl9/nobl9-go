package v1alphaExamples

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
	Object     any
	Variant    string
	SubVariant string
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
