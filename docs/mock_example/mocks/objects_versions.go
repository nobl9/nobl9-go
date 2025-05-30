// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/nobl9/nobl9-go/sdk/endpoints/objects (interfaces: Versions)
//
// Generated by this command:
//
//	mockgen -destination mocks/objects_versions.go -package mocks -mock_names Versions=MockObjectsVersions -typed github.com/nobl9/nobl9-go/sdk/endpoints/objects Versions
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

// MockObjectsVersions is a mock of Versions interface.
type MockObjectsVersions struct {
	ctrl     *gomock.Controller
	recorder *MockObjectsVersionsMockRecorder
	isgomock struct{}
}

// MockObjectsVersionsMockRecorder is the mock recorder for MockObjectsVersions.
type MockObjectsVersionsMockRecorder struct {
	mock *MockObjectsVersions
}

// NewMockObjectsVersions creates a new mock instance.
func NewMockObjectsVersions(ctrl *gomock.Controller) *MockObjectsVersions {
	mock := &MockObjectsVersions{ctrl: ctrl}
	mock.recorder = &MockObjectsVersionsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockObjectsVersions) EXPECT() *MockObjectsVersionsMockRecorder {
	return m.recorder
}

// V1 mocks base method.
func (m *MockObjectsVersions) V1() v1.Endpoints {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "V1")
	ret0, _ := ret[0].(v1.Endpoints)
	return ret0
}

// V1 indicates an expected call of V1.
func (mr *MockObjectsVersionsMockRecorder) V1() *MockObjectsVersionsV1Call {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "V1", reflect.TypeOf((*MockObjectsVersions)(nil).V1))
	return &MockObjectsVersionsV1Call{Call: call}
}

// MockObjectsVersionsV1Call wrap *gomock.Call
type MockObjectsVersionsV1Call struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockObjectsVersionsV1Call) Return(arg0 v1.Endpoints) *MockObjectsVersionsV1Call {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockObjectsVersionsV1Call) Do(f func() v1.Endpoints) *MockObjectsVersionsV1Call {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockObjectsVersionsV1Call) DoAndReturn(f func() v1.Endpoints) *MockObjectsVersionsV1Call {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
