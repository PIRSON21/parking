// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	models "github.com/PIRSON21/parking/internal/models"
	mock "github.com/stretchr/testify/mock"
)

// ParkingGetter is an autogenerated mock type for the ParkingGetter type
type ParkingGetter struct {
	mock.Mock
}

// GetParkingsList provides a mock function with given fields: search
func (_m *ParkingGetter) GetParkingsList(search string) ([]*models.Parking, error) {
	ret := _m.Called(search)

	if len(ret) == 0 {
		panic("no return value specified for GetParkingsList")
	}

	var r0 []*models.Parking
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]*models.Parking, error)); ok {
		return rf(search)
	}
	if rf, ok := ret.Get(0).(func(string) []*models.Parking); ok {
		r0 = rf(search)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Parking)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(search)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewParkingGetter creates a new instance of ParkingGetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewParkingGetter(t interface {
	mock.TestingT
	Cleanup(func())
}) *ParkingGetter {
	mock := &ParkingGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
