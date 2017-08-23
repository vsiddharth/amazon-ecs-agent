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
package platformResources

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/api"
	"github.com/aws/amazon-ecs-agent/agent/resources/cgroup/factory/mock"
	"github.com/aws/amazon-ecs-agent/agent/resources/cgroup/mock_control"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	testTaskArn = "arn:aws:ecs:region:account-id:task/task-id"
)

func TestSetupCgroupHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)
	mockCgroup := mock_cgroups.NewMockCgroup(ctrl)

	testTask := &api.Task{
		Arn: testTaskArn,
	}

	gomock.InOrder(
		mockControl.EXPECT().Exists(gomock.Any()).Return(false),
		mockControl.EXPECT().Create(gomock.Any()).Return(mockCgroup, nil),
	)

	platformResources := newPlatformResources(mockControl)
	assert.NoError(t, platformResources.Setup(testTask))
}

func TestSetupCgroupTasArnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)

	testTask := &api.Task{
		Arn: "invalid-task-ARN",
	}

	platformResources := newPlatformResources(mockControl)
	assert.Error(t, platformResources.Setup(testTask))
}

func TestSetupCgroupExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)

	testTask := &api.Task{
		Arn: testTaskArn,
	}

	mockControl.EXPECT().Exists(gomock.Any()).Return(true)

	platformResources := newPlatformResources(mockControl)
	assert.NoError(t, platformResources.Setup(testTask))
}

func TestSetupCgroupCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)

	testTask := &api.Task{
		Arn: testTaskArn,
	}

	gomock.InOrder(
		mockControl.EXPECT().Exists(gomock.Any()).Return(false),
		mockControl.EXPECT().Create(gomock.Any()).Return(nil, errors.New("cgroup create error")),
	)

	platformResources := newPlatformResources(mockControl)
	assert.Error(t, platformResources.Setup(testTask))
}

func TestSetupCgroupCreateNilError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)

	testTask := &api.Task{
		Arn: testTaskArn,
	}

	gomock.InOrder(
		mockControl.EXPECT().Exists(gomock.Any()).Return(false),
		mockControl.EXPECT().Create(gomock.Any()).Return(nil, nil),
	)

	platformResources := newPlatformResources(mockControl)
	assert.Error(t, platformResources.Setup(testTask))
}

func TestPlatformResourcesCleanupHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)

	testTask := &api.Task{
		Arn: testTaskArn,
	}

	mockControl.EXPECT().Remove(gomock.Any()).Return(nil)

	platformResources := newPlatformResources(mockControl)
	assert.NoError(t, platformResources.Cleanup(testTask))
}

func TestPlatformResourcesCleanupTaskARNError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)

	testTask := &api.Task{
		Arn: "invalid-task-ARN",
	}

	platformResources := newPlatformResources(mockControl)
	assert.Error(t, platformResources.Cleanup(testTask))
}

func TestPlatformResourcesCleanupRemoveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockControl := mock_cgroup.NewMockControl(ctrl)

	testTask := &api.Task{
		Arn: testTaskArn,
	}

	mockControl.EXPECT().Remove(gomock.Any()).Return(errors.New("cgroup remove error"))

	platformResources := newPlatformResources(mockControl)
	assert.Error(t, platformResources.Cleanup(testTask))
}
