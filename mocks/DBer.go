// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// DBer is an autogenerated mock type for the DBer type
type DBer struct {
	mock.Mock
}

// Acquire provides a mock function with given fields: _a0, _a1
func (_m *DBer) Acquire(_a0 string, _a1 time.Duration) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, time.Duration) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: _a0
func (_m *DBer) Delete(_a0 string) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
