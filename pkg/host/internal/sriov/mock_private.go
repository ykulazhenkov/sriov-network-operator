// Code generated by MockGen. DO NOT EDIT.
// Source: sriov.go

// Package sriov is a generated GoMock package.
package sriov

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
)

// MocksriovPrivateInterface is a mock of sriovPrivateInterface interface.
type MocksriovPrivateInterface struct {
	ctrl     *gomock.Controller
	recorder *MocksriovPrivateInterfaceMockRecorder
}

// MocksriovPrivateInterfaceMockRecorder is the mock recorder for MocksriovPrivateInterface.
type MocksriovPrivateInterfaceMockRecorder struct {
	mock *MocksriovPrivateInterface
}

// NewMocksriovPrivateInterface creates a new mock instance.
func NewMocksriovPrivateInterface(ctrl *gomock.Controller) *MocksriovPrivateInterface {
	mock := &MocksriovPrivateInterface{ctrl: ctrl}
	mock.recorder = &MocksriovPrivateInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocksriovPrivateInterface) EXPECT() *MocksriovPrivateInterfaceMockRecorder {
	return m.recorder
}

// addUdevRules mocks base method.
func (m *MocksriovPrivateInterface) addUdevRules(iface *v1.Interface) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "addUdevRules", iface)
	ret0, _ := ret[0].(error)
	return ret0
}

// addUdevRules indicates an expected call of addUdevRules.
func (mr *MocksriovPrivateInterfaceMockRecorder) addUdevRules(iface interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "addUdevRules", reflect.TypeOf((*MocksriovPrivateInterface)(nil).addUdevRules), iface)
}

// checkExternallyManagedPF mocks base method.
func (m *MocksriovPrivateInterface) checkExternallyManagedPF(iface *v1.Interface, ifaceStatus *v1.InterfaceExt) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "checkExternallyManagedPF", iface, ifaceStatus)
	ret0, _ := ret[0].(error)
	return ret0
}

// checkExternallyManagedPF indicates an expected call of checkExternallyManagedPF.
func (mr *MocksriovPrivateInterfaceMockRecorder) checkExternallyManagedPF(iface, ifaceStatus interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "checkExternallyManagedPF", reflect.TypeOf((*MocksriovPrivateInterface)(nil).checkExternallyManagedPF), iface, ifaceStatus)
}

// configSriovPFDevice mocks base method.
func (m *MocksriovPrivateInterface) configSriovPFDevice(iface *v1.Interface, ifaceStatus *v1.InterfaceExt) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "configSriovPFDevice", iface, ifaceStatus)
	ret0, _ := ret[0].(error)
	return ret0
}

// configSriovPFDevice indicates an expected call of configSriovPFDevice.
func (mr *MocksriovPrivateInterfaceMockRecorder) configSriovPFDevice(iface, ifaceStatus interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "configSriovPFDevice", reflect.TypeOf((*MocksriovPrivateInterface)(nil).configSriovPFDevice), iface, ifaceStatus)
}

// configSriovVFDevices mocks base method.
func (m *MocksriovPrivateInterface) configSriovVFDevices(iface *v1.Interface, ifaceStatus *v1.InterfaceExt) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "configSriovVFDevices", iface, ifaceStatus)
	ret0, _ := ret[0].(error)
	return ret0
}

// configSriovVFDevices indicates an expected call of configSriovVFDevices.
func (mr *MocksriovPrivateInterfaceMockRecorder) configSriovVFDevices(iface, ifaceStatus interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "configSriovVFDevices", reflect.TypeOf((*MocksriovPrivateInterface)(nil).configSriovVFDevices), iface, ifaceStatus)
}

// createSwitchdevVFs mocks base method.
func (m *MocksriovPrivateInterface) createSwitchdevVFs(iface *v1.Interface) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "createSwitchdevVFs", iface)
	ret0, _ := ret[0].(error)
	return ret0
}

// createSwitchdevVFs indicates an expected call of createSwitchdevVFs.
func (mr *MocksriovPrivateInterfaceMockRecorder) createSwitchdevVFs(iface interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "createSwitchdevVFs", reflect.TypeOf((*MocksriovPrivateInterface)(nil).createSwitchdevVFs), iface)
}

// createVFs mocks base method.
func (m *MocksriovPrivateInterface) createVFs(iface *v1.Interface, ifaceStatus *v1.InterfaceExt) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "createVFs", iface, ifaceStatus)
	ret0, _ := ret[0].(error)
	return ret0
}

// createVFs indicates an expected call of createVFs.
func (mr *MocksriovPrivateInterfaceMockRecorder) createVFs(iface, ifaceStatus interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "createVFs", reflect.TypeOf((*MocksriovPrivateInterface)(nil).createVFs), iface, ifaceStatus)
}

// removeUdevRules mocks base method.
func (m *MocksriovPrivateInterface) removeUdevRules(pciAddress string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "removeUdevRules", pciAddress)
	ret0, _ := ret[0].(error)
	return ret0
}

// removeUdevRules indicates an expected call of removeUdevRules.
func (mr *MocksriovPrivateInterfaceMockRecorder) removeUdevRules(pciAddress interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "removeUdevRules", reflect.TypeOf((*MocksriovPrivateInterface)(nil).removeUdevRules), pciAddress)
}

// unbindAllVFsOnPF mocks base method.
func (m *MocksriovPrivateInterface) unbindAllVFsOnPF(addr string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "unbindAllVFsOnPF", addr)
	ret0, _ := ret[0].(error)
	return ret0
}

// unbindAllVFsOnPF indicates an expected call of unbindAllVFsOnPF.
func (mr *MocksriovPrivateInterfaceMockRecorder) unbindAllVFsOnPF(addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "unbindAllVFsOnPF", reflect.TypeOf((*MocksriovPrivateInterface)(nil).unbindAllVFsOnPF), addr)
}
