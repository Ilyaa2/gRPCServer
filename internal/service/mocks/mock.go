// Code generated by MockGen. DO NOT EDIT.
// Source: service.go
//
// Generated by this command:
//
//	mockgen -source=service.go -destination=mocks/mock.go
//
// Package mock_service is a generated GoMock package.
package mock_service

import (
	context "context"
	dataModification "gRPCServer/internal/transport/grpc/sources/dataModification"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockEmployee is a mock of Employee interface.
type MockEmployee struct {
	ctrl     *gomock.Controller
	recorder *MockEmployeeMockRecorder
}

// MockEmployeeMockRecorder is the mock recorder for MockEmployee.
type MockEmployeeMockRecorder struct {
	mock *MockEmployee
}

// NewMockEmployee creates a new mock instance.
func NewMockEmployee(ctrl *gomock.Controller) *MockEmployee {
	mock := &MockEmployee{ctrl: ctrl}
	mock.recorder = &MockEmployeeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEmployee) EXPECT() *MockEmployeeMockRecorder {
	return m.recorder
}

// GetReasonOfAbsence mocks base method.
func (m *MockEmployee) GetReasonOfAbsence(ctx context.Context, details *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReasonOfAbsence", ctx, details)
	ret0, _ := ret[0].(*dataModification.ContactDetails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetReasonOfAbsence indicates an expected call of GetReasonOfAbsence.
func (mr *MockEmployeeMockRecorder) GetReasonOfAbsence(ctx, details any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReasonOfAbsence", reflect.TypeOf((*MockEmployee)(nil).GetReasonOfAbsence), ctx, details)
}
