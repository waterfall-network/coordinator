// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1 (interfaces: BeaconChainAltairServer)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	v2 "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

// MockBeaconChainAltairServer is a mock of BeaconChainAltairServer interface
type MockBeaconChainAltairServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChainAltairServerMockRecorder
}

// MockBeaconChainAltairServerMockRecorder is the mock recorder for MockBeaconChainAltairServer
type MockBeaconChainAltairServerMockRecorder struct {
	mock *MockBeaconChainAltairServer
}

// NewMockBeaconChainAltairServer creates a new mock instance
func NewMockBeaconChainAltairServer(ctrl *gomock.Controller) *MockBeaconChainAltairServer {
	mock := &MockBeaconChainAltairServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChainAltairServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChainAltairServer) EXPECT() *MockBeaconChainAltairServerMockRecorder {
	return m.recorder
}

// ListBlocks mocks base method
func (m *MockBeaconChainAltairServer) ListBlocks(arg0 context.Context, arg1 *v1alpha1.ListBlocksRequest) (*v2.ListBlocksResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListBlocks", arg0, arg1)
	ret0, _ := ret[0].(*v2.ListBlocksResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListBlocks indicates an expected call of ListBlocks
func (mr *MockBeaconChainAltairServerMockRecorder) ListBlocks(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListBlocks", reflect.TypeOf((*MockBeaconChainAltairServer)(nil).ListBlocks), arg0, arg1)
}
