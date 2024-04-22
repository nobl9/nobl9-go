package validation

import (
	"reflect"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// PropertyPlan is a validation plan for a single property.
type PropertyPlan struct {
	Path     string     `json:"path"`
	Type     string     `json:"type"`
	Examples []string   `json:"examples,omitempty"`
	Rules    []RulePlan `json:"rules,omitempty"`
}

// RulePlan is a validation plan for a single rule.
type RulePlan struct {
	Description string    `json:"description"`
	Details     string    `json:"details,omitempty"`
	ErrorCode   ErrorCode `json:"errorCode,omitempty"`
	Conditions  []string  `json:"conditions,omitempty"`
}

// Plan creates a validation plan for the provided [Validator].
// Each property is represented by a [PropertyPlan] which aggregates its every [RulePlan].
// If a property does not have any rules, it won't be included in the result.
func Plan[S any](v Validator[S]) []PropertyPlan {
	all := make([]planBuilder, 0)
	v.plan(planBuilder{path: "$", all: &all})
	propertiesMap := make(map[string]PropertyPlan)
	for _, p := range all {
		if p.rulePlan.Description == "" {
			p.rulePlan.Description = "TODO"
		}
		entry, ok := propertiesMap[p.path]
		if ok {
			entry.Rules = append(entry.Rules, p.rulePlan)
			propertiesMap[p.path] = entry
		} else {
			entry = PropertyPlan{
				Path:     p.path,
				Type:     p.propertyPlan.Type,
				Examples: p.propertyPlan.Examples,
				Rules:    []RulePlan{p.rulePlan},
			}
			propertiesMap[p.path] = entry
		}
	}
	properties := maps.Values(propertiesMap)
	sort.Slice(properties, func(i, j int) bool { return properties[i].Path < properties[j].Path })
	return properties
}

// planner is an interface for types that can create a [PropertyPlan] or [RulePlan].
type planner interface {
	plan(builder planBuilder)
}

// planBuilder is used to traverse the validation rules and build a slice of [PropertyPlan].
type planBuilder struct {
	path         string
	rulePlan     RulePlan
	propertyPlan PropertyPlan
	// all stores every rule for the current property.
	// It's not safe for concurrent usage.
	all *[]planBuilder
}

func (p planBuilder) append(path string) planBuilder {
	return planBuilder{
		path:         p.path + "." + path,
		all:          p.all,
		rulePlan:     p.rulePlan,
		propertyPlan: p.propertyPlan,
	}
}

// getTypeString returns the string representation of the type T.
// It returns the type name without package path or name.
// It strips the pointer '*' from the type name.
func getTypeString[T any]() string {
	typ := reflect.TypeOf(*new(T))
	if typ == nil {
		return ""
	}
	var result string
	if typ.PkgPath() == "" {
		result = typ.String()
	} else {
		result = typ.Name()
	}
	return strings.TrimPrefix(result, "*")
}