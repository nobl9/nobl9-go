package main

import (
	"fmt"
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

	doc := PropertyDoc{Path: path}
	doc = setTypeInfo(doc, typ)
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

func setTypeInfo(doc PropertyDoc, typ reflect.Type) PropertyDoc {
	info := extractTypeInfo(typ)
	doc.Type = info.Name
	doc.Package = info.Package
	return doc
}

// customTypesMapping is a re-mapping for custom typeInfo, useful for custom types like enums,
// which otherwise would be displayed as integers instead of strings.
var customTypesMapping = map[typeInfo]typeInfo{
	{Name: "Kind", Package: "github.com/nobl9/nobl9-go/manifest"}: {Name: "string"},
}

type typeInfo struct {
	Name    string
	Package string
}

func extractTypeInfo(typ reflect.Type) typeInfo {
	var info typeInfo
	if typ == nil {
		return info
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		info.Name = "[]"
	}
	switch typ.Kind() {
	case reflect.Map:
		key := extractTypeInfo(typ.Key())
		value := extractTypeInfo(typ.Elem())
		info.Name += fmt.Sprintf("map[%s]%s", key.Name, value.Name)
	case reflect.Struct:
		if typ.PkgPath() == "" {
			info.Name += typ.String()
		} else {
			info.Name += typ.Name()
			info.Package = typ.PkgPath()
		}
	default:
		var checkInfo typeInfo
		if typ.PkgPath() == "" {
			checkInfo.Name += typ.String()
		} else {
			checkInfo.Name += typ.Name()
			checkInfo.Package = typ.PkgPath()
		}
		if customType, ok := customTypesMapping[checkInfo]; ok {
			info = customType
		} else {
			info.Name += typ.Kind().String()
		}
	}
	return info
}
