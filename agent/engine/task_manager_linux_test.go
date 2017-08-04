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

package engine

import (
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/api"
	"github.com/stretchr/testify/assert"
)

// TestSetupCgroupWithInvalidSpec checks if managedTask can be setup with an
// invalid cgroup spec
func TestSetupCgroupWithInvalidSpec(t *testing.T) {
	mtask := managedTask{
		Task: &api.Task{},
	}

	err := mtask.SetupCgroup()
	assert.Error(t, err, "invalid cgroup spec")
}

// TestCleanupCgroupErrorPath attempts to check cgroup cleanup can be performed
// based off an erroneous spec
func TestCleanupCgroupErrorPath(t *testing.T) {
	mtask := managedTask{
		Task: &api.Task{},
	}

	err := mtask.CleanupCgroup()
	assert.Error(t, err)
}

// TODO: Add tests to cover happy paths
