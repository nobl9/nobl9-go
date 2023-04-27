// Package nobl9 provide an abstraction for communication with API
package nobl9

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// APIVersion is a value of valid apiVersions
const (
	APIVersion = "n9/v1alpha"
)

// HiddenValue can be used as a value of a secret field and is ignored during saving
const HiddenValue = "[hidden]"

// Possible values of field kind for valid Objects.
const (
	KindSLO         = "SLO"
	KindService     = "Service"
	KindAgent       = "Agent"
	KindProject     = "Project"
	KindAlertPolicy = "AlertPolicy"
	KindAlert       = "Alert"
	KindAlertMethod = "AlertMethod"
	KindDirect      = "Direct"
	KindDataExport  = "DataExport"
	KindGroup       = "Group"
	KindRoleBinding = "RoleBinding"
)

// APIObjects - all Objects available for this version of API
// Sorted in order of applying
type APIObjects struct {
	SLOs          []SLO
	Services      []Service
	Agents        []Agent
	AlertPolicies []AlertPolicy
	Alerts        []Alert
	AlertMethods  []AlertMethod
	Directs       []Direct
	DataExports   []DataExport
	Projects      []Project
	RoleBindings  []RoleBinding
	Groups        []Group
}

type Payload struct {
	objects []AnyJSONObj
}

func (p *Payload) AddObject(in interface{}) {
	p.objects = append(p.objects, toAnyJSONObj(in))
}

func (p *Payload) GetObjects() []AnyJSONObj {
	return p.objects
}

func toAnyJSONObj(in interface{}) AnyJSONObj {
	tmp, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	var out AnyJSONObj
	if err := json.Unmarshal(tmp, &out); err != nil {
		panic(err)
	}
	return out
}

// UnsupportedKindErr returns appropriate error for missing value in field kind
// for not empty field kind returns always that is not supported for this apiVersion
// so have to be validated before
func UnsupportedKindErr(o ObjectGeneric) error {
	if strings.TrimSpace(o.Kind) == "" {
		return EnhanceError(o, errors.New("missing or empty field kind for an Object"))
	}
	return EnhanceError(o, fmt.Errorf("invalid Object kind: %s for apiVersion: %s", o.Kind, o.APIVersion))
}

// ObjectInternal represents part of object which is only for internal usage,
// not exposed to the client, for internal usage
type ObjectInternal struct {
	Organization string `json:",omitempty" example:"nobl9-dev"`
	ManifestSrc  string `json:",omitempty" example:"x.yml"`
	OktaClientID string `json:"-"` // used only by kind Agent
}

// Metadata represents part of object which is common for all available Objects, for internal usage
type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Project     string `json:"project,omitempty"`
	Labels      Labels `json:"labels,omitempty"`
}

type Labels map[string][]string

// MetadataHolder is an intermediate structure that can provides metadata related
// field to other structures
type MetadataHolder struct {
	Metadata Metadata `json:"metadata"`
}

// ObjectHeader represents Header which is common for all available Objects
type ObjectHeader struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	MetadataHolder
	ObjectInternal
}

// ObjectGeneric represents struct to which every Object is parsable
// Specific types of Object have different structures as Spec
type ObjectGeneric struct {
	ObjectHeader
	Spec json.RawMessage `json:"spec"`
}

type HistoricalDataRetrieval struct {
	DefaultDuration HistoricalDataRetrievalDuration `json:"defaultDuration"`
	MaxDuration     HistoricalDataRetrievalDuration `json:"maxDuration"`
}

type HistoricalDataRetrievalDuration struct {
	Unit  string      `json:"unit" example:"Day"`
	Value json.Number `json:"value" example:"30"`
}

type QueryDelayDuration struct {
	Unit  string      `json:"unit" example:"Minute"`
	Value json.Number `json:"value" example:"1"`
}

// EnhanceError annotates error with path of manifest source, if it exists
// it not returns the same error as passed as argument
func EnhanceError(o ObjectGeneric, err error) error {
	if err != nil && o.ManifestSrc != "" {
		err = fmt.Errorf("%s:\n%w", o.ManifestSrc, err)
	}
	return err
}

// Parse takes care of all Object supported by n9/v1alpha apiVersion
func Parse(o ObjectGeneric, parsedObjects *APIObjects, onlyHeaders bool) error {
	var allErrors []string
	switch o.Kind {
	case KindSLO:
		slo, err := genericToSLO(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.SLOs = append(parsedObjects.SLOs, slo)
	case KindService:
		service, err := genericToService(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.Services = append(parsedObjects.Services, service)
	case KindAgent:
		agent, err := genericToAgent(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.Agents = append(parsedObjects.Agents, agent)
	case KindAlertPolicy:
		alertPolicy, err := genericToAlertPolicy(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.AlertPolicies = append(parsedObjects.AlertPolicies, alertPolicy)
	case KindAlertMethod:
		alertMethod, err := genericToAlertMethod(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.AlertMethods = append(parsedObjects.AlertMethods, alertMethod)
	case KindDirect:
		direct, err := genericToDirect(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.Directs = append(parsedObjects.Directs, direct)
	case KindDataExport:
		dataExport, err := genericToDataExport(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.DataExports = append(parsedObjects.DataExports, dataExport)
	case KindProject:
		project, err := genericToProject(o, onlyHeaders)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.Projects = append(parsedObjects.Projects, project)
	case KindRoleBinding:
		roleBinding, err := genericToRoleBinding(o)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.RoleBindings = append(parsedObjects.RoleBindings, roleBinding)
	case KindGroup:
		group, err := genericToGroup(o)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		parsedObjects.Groups = append(parsedObjects.Groups, group)
	// catching invalid kinds of objects for this apiVersion
	default:
		err := UnsupportedKindErr(o)
		allErrors = append(allErrors, err.Error())
	}
	if len(allErrors) > 0 {
		return errors.New(strings.Join(allErrors, "\n"))
	}
	return nil
}
