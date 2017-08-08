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
	"github.com/aws/amazon-ecs-agent/agent/resources/cgroup"
	"github.com/cihub/seelog"
	"github.com/pkg/errors"
)

// SetupPlatformResources sets up platform level resources
func (mtask *managedTask) SetupPlatformResources() error {
	if mtask.Task.CgroupEnabled() {
		return mtask.setupCgroup()
	}
	return nil
}

// CleanupPlatformResources cleans up platform level resources
func (mtask *managedTask) CleanupPlatformResources() error {
	if mtask.Task.CgroupEnabled() {
		return mtask.cleanupCgroup()
	}
	return nil
}

// setupCgroup sets up the cgroup for each managed task
func (mtask *managedTask) setupCgroup() error {
	// Grab cgroup spec
	cgroupSpec, err := mtask.Task.GetCgroupSpec()
	if err != nil {
		return errors.Wrapf(err, "cgroup setup: unable to obtain cgroup spec")
	}
	seelog.Debugf("Setting up task cgroup %s for task %s", cgroupSpec.Root, mtask.Task.Arn)

	// Create cgroup
	err = cgroup.Create(&cgroupSpec)
	if err != nil {
		return errors.Wrapf(err, "cgroup setup: unable to create")
	}
	return nil
}

// cleanupCgroup removes the task cgroup
func (mtask *managedTask) cleanupCgroup() error {
	// Grab cgroup spec
	cgroupSpec, err := mtask.Task.GetCgroupSpec()
	if err != nil {
		return errors.Wrapf(err, "cgroup cleanup: unable to obtain cgroup spec")
	}
	seelog.Debugf("Cleaning up task cgroup %s for task %s", cgroupSpec.Root, mtask.Task.Arn)

	return cgroup.Remove(&cgroupSpec)
}
