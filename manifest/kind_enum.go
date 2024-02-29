// Code generated by go-enum DO NOT EDIT.
// Version: 0.6.0
// Revision: 919e61c0174b91303753ee3898569a01abb32c97
// Build Date: 2023-12-18T15:54:43Z
// Built By: goreleaser

package manifest

import (
	"fmt"
	"strings"
)

const (
	// KindSLO is a Kind of type SLO.
	KindSLO Kind = iota + 1
	// KindService is a Kind of type Service.
	KindService
	// KindAgent is a Kind of type Agent.
	KindAgent
	// KindAlertPolicy is a Kind of type AlertPolicy.
	KindAlertPolicy
	// KindAlertSilence is a Kind of type AlertSilence.
	KindAlertSilence
	// KindAlert is a Kind of type Alert.
	KindAlert
	// KindProject is a Kind of type Project.
	KindProject
	// KindAlertMethod is a Kind of type AlertMethod.
	KindAlertMethod
	// KindDirect is a Kind of type Direct.
	KindDirect
	// KindDataExport is a Kind of type DataExport.
	KindDataExport
	// KindRoleBinding is a Kind of type RoleBinding.
	KindRoleBinding
	// KindAnnotation is a Kind of type Annotation.
	KindAnnotation
	// KindUserGroup is a Kind of type UserGroup.
	KindUserGroup
	// KindBudgetAdjustment is a Kind of type BudgetAdjustment.
	KindBudgetAdjustment
)

var ErrInvalidKind = fmt.Errorf("not a valid Kind, try [%s]", strings.Join(_KindNames, ", "))

const _KindName = "SLOServiceAgentAlertPolicyAlertSilenceAlertProjectAlertMethodDirectDataExportRoleBindingAnnotationUserGroupBudgetAdjustment"

var _KindNames = []string{
	_KindName[0:3],
	_KindName[3:10],
	_KindName[10:15],
	_KindName[15:26],
	_KindName[26:38],
	_KindName[38:43],
	_KindName[43:50],
	_KindName[50:61],
	_KindName[61:67],
	_KindName[67:77],
	_KindName[77:88],
	_KindName[88:98],
	_KindName[98:107],
	_KindName[107:123],
}

// KindNames returns a list of possible string values of Kind.
func KindNames() []string {
	tmp := make([]string, len(_KindNames))
	copy(tmp, _KindNames)
	return tmp
}

// KindValues returns a list of the values for Kind
func KindValues() []Kind {
	return []Kind{
		KindSLO,
		KindService,
		KindAgent,
		KindAlertPolicy,
		KindAlertSilence,
		KindAlert,
		KindProject,
		KindAlertMethod,
		KindDirect,
		KindDataExport,
		KindRoleBinding,
		KindAnnotation,
		KindUserGroup,
		KindBudgetAdjustment,
	}
}

var _KindMap = map[Kind]string{
	KindSLO:              _KindName[0:3],
	KindService:          _KindName[3:10],
	KindAgent:            _KindName[10:15],
	KindAlertPolicy:      _KindName[15:26],
	KindAlertSilence:     _KindName[26:38],
	KindAlert:            _KindName[38:43],
	KindProject:          _KindName[43:50],
	KindAlertMethod:      _KindName[50:61],
	KindDirect:           _KindName[61:67],
	KindDataExport:       _KindName[67:77],
	KindRoleBinding:      _KindName[77:88],
	KindAnnotation:       _KindName[88:98],
	KindUserGroup:        _KindName[98:107],
	KindBudgetAdjustment: _KindName[107:123],
}

// String implements the Stringer interface.
func (x Kind) String() string {
	if str, ok := _KindMap[x]; ok {
		return str
	}
	return fmt.Sprintf("Kind(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x Kind) IsValid() bool {
	_, ok := _KindMap[x]
	return ok
}

var _KindValue = map[string]Kind{
	_KindName[0:3]:                      KindSLO,
	strings.ToLower(_KindName[0:3]):     KindSLO,
	_KindName[3:10]:                     KindService,
	strings.ToLower(_KindName[3:10]):    KindService,
	_KindName[10:15]:                    KindAgent,
	strings.ToLower(_KindName[10:15]):   KindAgent,
	_KindName[15:26]:                    KindAlertPolicy,
	strings.ToLower(_KindName[15:26]):   KindAlertPolicy,
	_KindName[26:38]:                    KindAlertSilence,
	strings.ToLower(_KindName[26:38]):   KindAlertSilence,
	_KindName[38:43]:                    KindAlert,
	strings.ToLower(_KindName[38:43]):   KindAlert,
	_KindName[43:50]:                    KindProject,
	strings.ToLower(_KindName[43:50]):   KindProject,
	_KindName[50:61]:                    KindAlertMethod,
	strings.ToLower(_KindName[50:61]):   KindAlertMethod,
	_KindName[61:67]:                    KindDirect,
	strings.ToLower(_KindName[61:67]):   KindDirect,
	_KindName[67:77]:                    KindDataExport,
	strings.ToLower(_KindName[67:77]):   KindDataExport,
	_KindName[77:88]:                    KindRoleBinding,
	strings.ToLower(_KindName[77:88]):   KindRoleBinding,
	_KindName[88:98]:                    KindAnnotation,
	strings.ToLower(_KindName[88:98]):   KindAnnotation,
	_KindName[98:107]:                   KindUserGroup,
	strings.ToLower(_KindName[98:107]):  KindUserGroup,
	_KindName[107:123]:                  KindBudgetAdjustment,
	strings.ToLower(_KindName[107:123]): KindBudgetAdjustment,
}

// ParseKind attempts to convert a string to a Kind.
func ParseKind(name string) (Kind, error) {
	if x, ok := _KindValue[name]; ok {
		return x, nil
	}
	// Case insensitive parse, do a separate lookup to prevent unnecessary cost of lowercasing a string if we don't need to.
	if x, ok := _KindValue[strings.ToLower(name)]; ok {
		return x, nil
	}
	return Kind(0), fmt.Errorf("%s is %w", name, ErrInvalidKind)
}
