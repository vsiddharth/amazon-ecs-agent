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

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
)

func TestValidateCgroupSpecWithEmptySpec(t *testing.T) {
	err := validateCgroupSpec(nil)
	assert.Error(t, err, "empty cgroup spec")
}

func TestValidateCgroupSpecWithMissingRoot(t *testing.T) {
	cgroupSpec := Spec{}
	err := validateCgroupSpec(&cgroupSpec)
	assert.Error(t, err, "missing cgroup root")
}

func TestValidateCgroupSpecWithNonECSPrefix(t *testing.T) {
	cgroupSpec := Spec{
		Root: "/non-ecs/root",
	}

	err := validateCgroupSpec(&cgroupSpec)
	assert.Error(t, err, "cgroup root missing ecs prefix")
}

func TestValidateCgroupSpecWithMissingResourceSpecs(t *testing.T) {
	cgroupSpec := Spec{
		Root:  "/ecs/task-id",
		Specs: nil,
	}
	err := validateCgroupSpec(&cgroupSpec)
	assert.Error(t, err, "cgroup spec missing resource specs")
}

func TestValidateCgroupSpecWithHappyPath(t *testing.T) {
	cgroupSpec := Spec{
		Root:  "/ecs/task-id",
		Specs: &specs.LinuxResources{},
	}
	err := validateCgroupSpec(&cgroupSpec)
	assert.NoError(t, err, "happy path")
}

func TestCreateWithInvalidSpec(t *testing.T) {
	invalidCgroupSpec := Spec{
		Root:  "/ecs/task-id",
		Specs: nil,
	}
	err := Create(&invalidCgroupSpec)
	assert.Error(t, err, "invalid cgroup spec")
}

func TestLoadWithInvalidSpec(t *testing.T) {
	invalidCgroupSpec := Spec{
		Root:  "/ecs/task-id",
		Specs: nil,
	}
	_, err := load(&invalidCgroupSpec)
	assert.Error(t, err, "invalid cgroup spec")
}

func TestRemoveWithInvalidSpec(t *testing.T) {
	invalidCgroupSpec := Spec{
		Root:  "/ecs/task-id",
		Specs: nil,
	}
	err := Remove(&invalidCgroupSpec)
	assert.Error(t, err, "invalid cgroup spec")

}
