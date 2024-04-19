package validation

import (
	"sort"

	"golang.org/x/exp/maps"
)

type planner interface {
	plan(builder planBuilder)
}

type planBuilder struct {
	path         string
	rulePlan     RulePlan
	propertyPlan PropertyPlan
	all          *[]planBuilder
}

type RulePlan struct {
	Description string    `json:"description"`
	Details     string    `json:"details,omitempty"`
	ErrorCode   ErrorCode `json:"errorCode,omitempty"`
	Conditions  []string  `json:"conditions,omitempty"`
}

type PropertyPlan struct {
	Path     string     `json:"path"`
	Type     string     `json:"type"`
	Examples []string   `json:"examples,omitempty"`
	Rules    []RulePlan `json:"rules,omitempty"`
}

func (p planBuilder) append(path string) planBuilder {
	return planBuilder{
		path:         p.path + "." + path,
		all:          p.all,
		rulePlan:     p.rulePlan,
		propertyPlan: p.propertyPlan,
	}
}

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
