// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/repositories/job_repository.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	entities "gophermart/internal/app/entities"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pgx "github.com/jackc/pgx/v5"
)

// MockJobRepositoryInterface is a mock of JobRepositoryInterface interface.
type MockJobRepositoryInterface struct {
	ctrl     *gomock.Controller
	recorder *MockJobRepositoryInterfaceMockRecorder
}

// MockJobRepositoryInterfaceMockRecorder is the mock recorder for MockJobRepositoryInterface.
type MockJobRepositoryInterfaceMockRecorder struct {
	mock *MockJobRepositoryInterface
}

// NewMockJobRepositoryInterface creates a new mock instance.
func NewMockJobRepositoryInterface(ctrl *gomock.Controller) *MockJobRepositoryInterface {
	mock := &MockJobRepositoryInterface{ctrl: ctrl}
	mock.recorder = &MockJobRepositoryInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJobRepositoryInterface) EXPECT() *MockJobRepositoryInterfaceMockRecorder {
	return m.recorder
}

// DeleteJobByID mocks base method.
func (m *MockJobRepositoryInterface) DeleteJobByID(ctx context.Context, tx pgx.Tx, jobID int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJobByID", ctx, tx, jobID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJobByID indicates an expected call of DeleteJobByID.
func (mr *MockJobRepositoryInterfaceMockRecorder) DeleteJobByID(ctx, tx, jobID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJobByID", reflect.TypeOf((*MockJobRepositoryInterface)(nil).DeleteJobByID), ctx, tx, jobID)
}

// GetPendingJobs mocks base method.
func (m *MockJobRepositoryInterface) GetPendingJobs(ctx context.Context, tx pgx.Tx, limit int) ([]entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPendingJobs", ctx, tx, limit)
	ret0, _ := ret[0].([]entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPendingJobs indicates an expected call of GetPendingJobs.
func (mr *MockJobRepositoryInterfaceMockRecorder) GetPendingJobs(ctx, tx, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPendingJobs", reflect.TypeOf((*MockJobRepositoryInterface)(nil).GetPendingJobs), ctx, tx, limit)
}

// SaveJob mocks base method.
func (m *MockJobRepositoryInterface) SaveJob(ctx context.Context, tx pgx.Tx, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveJob", ctx, tx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveJob indicates an expected call of SaveJob.
func (mr *MockJobRepositoryInterfaceMockRecorder) SaveJob(ctx, tx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveJob", reflect.TypeOf((*MockJobRepositoryInterface)(nil).SaveJob), ctx, tx, job)
}

// UpdateJobPoolAt mocks base method.
func (m *MockJobRepositoryInterface) UpdateJobPoolAt(ctx context.Context, tx pgx.Tx, jobID int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobPoolAt", ctx, tx, jobID)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobPoolAt indicates an expected call of UpdateJobPoolAt.
func (mr *MockJobRepositoryInterfaceMockRecorder) UpdateJobPoolAt(ctx, tx, jobID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobPoolAt", reflect.TypeOf((*MockJobRepositoryInterface)(nil).UpdateJobPoolAt), ctx, tx, jobID)
}
