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
	"strings"

	"github.com/aws/amazon-ecs-agent/agent/api"
	"github.com/aws/amazon-ecs-agent/agent/config"
	"github.com/aws/amazon-ecs-agent/agent/resources/cgroup"
	"github.com/cihub/seelog"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

const (
	sepForwardSlash = "/"
)

// platformResources to abstract task platform resources
// Currently a composite type to track platform specific resources
type platformResources struct {
	control cgroup.Control
}

// New returns a new platformResources object
func New() PlatformResources {
	return newPlatformResources(cgroup.New())
}

// newPlatformResources helps setup the platformResources
func newPlatformResources(control cgroup.Control) PlatformResources {
	return &platformResources{
		control: control,
	}
}

// Setup helps setup the platform resources
func (p *platformResources) Setup(task *api.Task) error {
	return p.setupCgroup(task)
}

// setupCgroup sets up the task cgroup
func (p *platformResources) setupCgroup(task *api.Task) error {
	// Fetch taskID
	taskID, err := task.GetID()
	if err != nil {
		return errors.Wrapf(err, "platform resources: cgroup setup: unable to obtain taskID")
	}

	// Build cgroup root
	cgroupRoot := strings.Join([]string{config.DefaultTaskCgroupPrefix, taskID}, sepForwardSlash)

	// Check if cgroup already exists
	if p.control.Exists(cgroupRoot) {
		seelog.Debugf("platform resources: cgroup setup: already exists at: %s, skipping creation", cgroupRoot)
		return nil
	}

	// TODO: Build linux resources
	linuxResources := specs.LinuxResources{}

	// Populate cgroup spec
	cgroupSpec := cgroup.Spec{
		Root:  cgroupRoot,
		Specs: &linuxResources,
	}

	// Create cgroup
	cgrp, err := p.control.Create(&cgroupSpec)
	if err != nil {
		return errors.Wrapf(err, "platform resources setup: unable to create cgroup")
	}

	// FIXME: Should never happen
	if cgrp == nil {
		return errors.New("platform resources setup: invalid cgroup")
	}

	return nil
}

// Cleanup helps cleanup the task platform resources
func (p *platformResources) Cleanup(task *api.Task) error {
	return p.cleanupCgroup(task)
}

// cleanupCgroup removes the task cgroup
func (p *platformResources) cleanupCgroup(task *api.Task) error {
	taskID, err := task.GetID()
	if err != nil {
		return errors.Wrapf(err, "platform resources cleanup: unable to obtain taskID")
	}

	cgroupRoot := strings.Join([]string{config.DefaultTaskCgroupPrefix, taskID}, sepForwardSlash)

	err = p.control.Remove(cgroupRoot)
	if err != nil {
		return errors.Wrapf(err, "platform resources cleanup: unable to delete cgroup")
	}

	return nil
}
