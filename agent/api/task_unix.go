// +build !windows

// Copyright 2014-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package api

import (
	"github.com/aws/amazon-ecs-agent/agent/resources/cgroup"
	docker "github.com/fsouza/go-dockerclient"

	"github.com/pkg/errors"
)

const (
	portBindingHostIP = "0.0.0.0"
)

func (task *Task) adjustForPlatform() {}

func getCanonicalPath(path string) string { return path }

// GetCgroupSpec fetches the task cgroup spec
func (task *Task) GetCgroupSpec() (cgroup.Spec, error) {
	task.cgroupSpecLock.RLock()
	defer task.cgroupSpecLock.RUnlock()

	if task.CgroupSpec == nil {
		return cgroup.Spec{}, errors.New("task cgroup: missing spec")
	}

	return *task.CgroupSpec, nil
}

// updateHostConfigWithCgroupParent sets the cgroup parent for containers
func (task *Task) updateHostConfigWithCgroupParent(hostConfig *docker.HostConfig) error {
	// Get cgroup spec
	cgroupSpec, err := task.GetCgroupSpec()
	if err != nil {
		return errors.Wrapf(err, "task set cgroup parent: unable to get valid cgroup spec")
	}

	// Check for empty cgroup root
	if cgroupSpec.Root == "" {
		return errors.New("task set cgroup parent: empty cgroup root")
	}

	// Set cgroup parent
	hostConfig.CgroupParent = cgroupSpec.Root

	return nil
}
