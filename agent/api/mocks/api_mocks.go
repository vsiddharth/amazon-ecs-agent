// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/aws/amazon-ecs-agent/agent/api (interfaces: ECSSDK,ECSSubmitStateSDK,ECSClient)

package mock_api

import (
	api "github.com/aws/amazon-ecs-agent/agent/api"
	ecs "github.com/aws/amazon-ecs-agent/agent/ecs_client/model/ecs"
	gomock "github.com/golang/mock/gomock"
)

// Mock of ECSSDK interface
type MockECSSDK struct {
	ctrl     *gomock.Controller
	recorder *_MockECSSDKRecorder
}

// Recorder for MockECSSDK (not exported)
type _MockECSSDKRecorder struct {
	mock *MockECSSDK
}

func NewMockECSSDK(ctrl *gomock.Controller) *MockECSSDK {
	mock := &MockECSSDK{ctrl: ctrl}
	mock.recorder = &_MockECSSDKRecorder{mock}
	return mock
}

func (_m *MockECSSDK) EXPECT() *_MockECSSDKRecorder {
	return _m.recorder
}

func (_m *MockECSSDK) CreateCluster(_param0 *ecs.CreateClusterInput) (*ecs.CreateClusterOutput, error) {
	ret := _m.ctrl.Call(_m, "CreateCluster", _param0)
	ret0, _ := ret[0].(*ecs.CreateClusterOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSSDKRecorder) CreateCluster(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateCluster", arg0)
}

func (_m *MockECSSDK) DiscoverPollEndpoint(_param0 *ecs.DiscoverPollEndpointInput) (*ecs.DiscoverPollEndpointOutput, error) {
	ret := _m.ctrl.Call(_m, "DiscoverPollEndpoint", _param0)
	ret0, _ := ret[0].(*ecs.DiscoverPollEndpointOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSSDKRecorder) DiscoverPollEndpoint(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DiscoverPollEndpoint", arg0)
}

func (_m *MockECSSDK) RegisterContainerInstance(_param0 *ecs.RegisterContainerInstanceInput) (*ecs.RegisterContainerInstanceOutput, error) {
	ret := _m.ctrl.Call(_m, "RegisterContainerInstance", _param0)
	ret0, _ := ret[0].(*ecs.RegisterContainerInstanceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSSDKRecorder) RegisterContainerInstance(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RegisterContainerInstance", arg0)
}

// Mock of ECSSubmitStateSDK interface
type MockECSSubmitStateSDK struct {
	ctrl     *gomock.Controller
	recorder *_MockECSSubmitStateSDKRecorder
}

// Recorder for MockECSSubmitStateSDK (not exported)
type _MockECSSubmitStateSDKRecorder struct {
	mock *MockECSSubmitStateSDK
}

func NewMockECSSubmitStateSDK(ctrl *gomock.Controller) *MockECSSubmitStateSDK {
	mock := &MockECSSubmitStateSDK{ctrl: ctrl}
	mock.recorder = &_MockECSSubmitStateSDKRecorder{mock}
	return mock
}

func (_m *MockECSSubmitStateSDK) EXPECT() *_MockECSSubmitStateSDKRecorder {
	return _m.recorder
}

func (_m *MockECSSubmitStateSDK) SubmitContainerStateChange(_param0 *ecs.SubmitContainerStateChangeInput) (*ecs.SubmitContainerStateChangeOutput, error) {
	ret := _m.ctrl.Call(_m, "SubmitContainerStateChange", _param0)
	ret0, _ := ret[0].(*ecs.SubmitContainerStateChangeOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSSubmitStateSDKRecorder) SubmitContainerStateChange(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SubmitContainerStateChange", arg0)
}

func (_m *MockECSSubmitStateSDK) SubmitTaskStateChange(_param0 *ecs.SubmitTaskStateChangeInput) (*ecs.SubmitTaskStateChangeOutput, error) {
	ret := _m.ctrl.Call(_m, "SubmitTaskStateChange", _param0)
	ret0, _ := ret[0].(*ecs.SubmitTaskStateChangeOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSSubmitStateSDKRecorder) SubmitTaskStateChange(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SubmitTaskStateChange", arg0)
}

// Mock of ECSClient interface
type MockECSClient struct {
	ctrl     *gomock.Controller
	recorder *_MockECSClientRecorder
}

// Recorder for MockECSClient (not exported)
type _MockECSClientRecorder struct {
	mock *MockECSClient
}

func NewMockECSClient(ctrl *gomock.Controller) *MockECSClient {
	mock := &MockECSClient{ctrl: ctrl}
	mock.recorder = &_MockECSClientRecorder{mock}
	return mock
}

func (_m *MockECSClient) EXPECT() *_MockECSClientRecorder {
	return _m.recorder
}

func (_m *MockECSClient) DiscoverPollEndpoint(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "DiscoverPollEndpoint", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSClientRecorder) DiscoverPollEndpoint(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DiscoverPollEndpoint", arg0)
}

func (_m *MockECSClient) DiscoverTelemetryEndpoint(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "DiscoverTelemetryEndpoint", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSClientRecorder) DiscoverTelemetryEndpoint(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DiscoverTelemetryEndpoint", arg0)
}

func (_m *MockECSClient) RegisterContainerInstance(_param0 string, _param1 []*ecs.Attribute) (string, error) {
	ret := _m.ctrl.Call(_m, "RegisterContainerInstance", _param0, _param1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockECSClientRecorder) RegisterContainerInstance(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RegisterContainerInstance", arg0, arg1)
}

func (_m *MockECSClient) SubmitContainerStateChange(_param0 api.ContainerStateChange) error {
	ret := _m.ctrl.Call(_m, "SubmitContainerStateChange", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockECSClientRecorder) SubmitContainerStateChange(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SubmitContainerStateChange", arg0)
}

func (_m *MockECSClient) SubmitTaskStateChange(_param0 api.TaskStateChange) error {
	ret := _m.ctrl.Call(_m, "SubmitTaskStateChange", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockECSClientRecorder) SubmitTaskStateChange(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SubmitTaskStateChange", arg0)
}
