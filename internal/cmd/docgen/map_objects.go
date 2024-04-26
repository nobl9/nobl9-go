package main

import (
	"reflect"
	"strings"
)

func newObjectMapper() *objectMapper {
	return &objectMapper{}
}

type objectMapper struct {
	Properties []PropertyDoc
}

func (o *objectMapper) Map(typ reflect.Type, path string) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	doc := PropertyDoc{
		Path:    path,
		Type:    typ.Name(),
		Package: typ.PkgPath(),
	}
	if doc.Package == "" {
		doc.Type = typ.String()
	}
	o.Properties = append(o.Properties, doc)

	switch typ.Kind() {
	case reflect.Struct:
		for _, field := range reflect.VisibleFields(typ) {
			tags := strings.Split(field.Tag.Get("json"), ",")
			if len(tags) == 0 {
				continue
			}
			name := tags[0]
			if name == "" || name == "-" {
				continue
			}
			o.Map(field.Type, path+"."+name)
		}
	case reflect.Slice:
		o.Map(typ.Elem(), path+"[*]")
	case reflect.Map:
		o.Map(typ.Elem(), path+".*")
	default:
	}
}
