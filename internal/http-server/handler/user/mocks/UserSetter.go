// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	models "github.com/PIRSON21/parking/internal/models"
	mock "github.com/stretchr/testify/mock"
)

// UserSetter is an autogenerated mock type for the UserSetter type
type UserSetter struct {
	mock.Mock
}

// CreateNewManager provides a mock function with given fields: _a0
func (_m *UserSetter) CreateNewManager(_a0 *models.User) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for CreateNewManager")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.User) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewUserSetter creates a new instance of UserSetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserSetter(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserSetter {
	mock := &UserSetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
