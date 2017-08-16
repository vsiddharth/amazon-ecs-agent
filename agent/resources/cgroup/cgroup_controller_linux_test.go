// +build linux

// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package cgroup

import (
	"testing"

	"errors"

	"github.com/aws/amazon-ecs-agent/agent/resources/cgroup/mock"
	"github.com/containerd/cgroups"
	"github.com/golang/mock/gomock"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
)

func TestCreateHappyCase(t *testing.T) {
	ctrl, testGroup, _ := setupMocks(t)
	defer ctrl.Finish()

	testString := "/ecs/foo"
	testSpecs := &specs.LinuxResources{}

	res, err := Create(&Spec{testString, testSpecs})
	assert.Equal(t, testGroup, res)
	assert.NoError(t, err)
}

func TestCreateErrorCase(t *testing.T) {
	ctrl, _, mockFactory := setupMocks(t)
	defer ctrl.Finish()

	mockFactory.err = errors.New("cgroup exploded")

	res, err := Create(&Spec{"/ecs/foo", &specs.LinuxResources{}})
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreateWithBadSpecs(t *testing.T) {
	ctrl, _, _ := setupMocks(t)
	defer ctrl.Finish()

	var nil_string string

	testCases := []struct {
		spec *Spec
		name string
	}{
		{&Spec{"", nil}, "empty root and nil spec"},
		{&Spec{"/ecs/foo", nil}, "root with nil spec"},
		{&Spec{"", &specs.LinuxResources{}}, "empty root with spec"},
		{&Spec{}, "empty spec"},
		{&Spec{nil_string, &specs.LinuxResources{}}, "nil root with spec"},
		{nil, "nil spec"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			control, err := Create(tc.spec)
			assert.Error(t, err, "Create should return an error")
			assert.Nil(t, control, "Create call should not return a controller")
		})
	}
}

func TestRemoveHappyCase(t *testing.T) {
	ctrl, testGroup, _ := setupMocks(t)
	defer ctrl.Finish()

	testGroup.EXPECT().Delete().Times(1)
	err := Remove("/ecs/foo")
	assert.NoError(t, err)
}

func TestRemoveDoingErrorCase(t *testing.T) {
	ctrl, _, mockFactory := setupMocks(t)
	defer ctrl.Finish()

	mockFactory.group = nil
	mockFactory.err = errors.New("Unable to load")

	err := Remove("/ecs/foo")
	assert.Error(t, err)
}

func TestRemoveErrorCase(t *testing.T) {
	ctrl, testGroup, _ := setupMocks(t)
	defer ctrl.Finish()

	testGroup.EXPECT().Delete().Times(1).Return(errors.New("Cgroup error"))
	err := Remove("/ecs/foo")
	assert.Error(t, err)
}

func setupMocks(t *testing.T) (*gomock.Controller, *mock_cgroups.MockCgroup, *mockCgroupFactory) {
	ctrl := gomock.NewController(t)
	testGroup := mock_cgroups.NewMockCgroup(ctrl)
	mockFactory := &mockCgroupFactory{testGroup, nil}

	factory = mockFactory
	return ctrl, testGroup, mockFactory
}

type mockCgroupFactory struct {
	group cgroups.Cgroup
	err   error
}

func (f *mockCgroupFactory) New(hierarchy cgroups.Hierarchy, path cgroups.Path, specs *specs.LinuxResources) (cgroups.Cgroup, error) {
	return f.group, f.err
}

func (f *mockCgroupFactory) Load(hierarchy cgroups.Hierarchy, path cgroups.Path) (cgroups.Cgroup, error) {
	return f.group, f.err
}
