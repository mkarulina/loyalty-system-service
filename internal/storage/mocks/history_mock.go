// Code generated by MockGen. DO NOT EDIT.
// Source: internal/storage/history.go

// Package mock_storage is a generated GoMock package.
package mock_storage

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	storage "github.com/mkarulina/loyalty-system-service.git/internal/storage"
)

// MockHistoryStorage is a mock of HistoryStorage interface.
type MockHistoryStorage struct {
	ctrl     *gomock.Controller
	recorder *MockHistoryStorageMockRecorder
}

// MockHistoryStorageMockRecorder is the mock recorder for MockHistoryStorage.
type MockHistoryStorageMockRecorder struct {
	mock *MockHistoryStorage
}

// NewMockHistoryStorage creates a new mock instance.
func NewMockHistoryStorage(ctrl *gomock.Controller) *MockHistoryStorage {
	mock := &MockHistoryStorage{ctrl: ctrl}
	mock.recorder = &MockHistoryStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHistoryStorage) EXPECT() *MockHistoryStorageMockRecorder {
	return m.recorder
}

// AddWithdrawnHistory mocks base method.
func (m *MockHistoryStorage) AddWithdrawnHistory(user, order string, sum float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddWithdrawnHistory", user, order, sum)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddWithdrawnHistory indicates an expected call of AddWithdrawnHistory.
func (mr *MockHistoryStorageMockRecorder) AddWithdrawnHistory(user, order, sum interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddWithdrawnHistory", reflect.TypeOf((*MockHistoryStorage)(nil).AddWithdrawnHistory), user, order, sum)
}

// GetWithdrawalsHistory mocks base method.
func (m *MockHistoryStorage) GetWithdrawalsHistory(token string) ([]storage.Withdrawn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithdrawalsHistory", token)
	ret0, _ := ret[0].([]storage.Withdrawn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithdrawalsHistory indicates an expected call of GetWithdrawalsHistory.
func (mr *MockHistoryStorageMockRecorder) GetWithdrawalsHistory(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithdrawalsHistory", reflect.TypeOf((*MockHistoryStorage)(nil).GetWithdrawalsHistory), token)
}
