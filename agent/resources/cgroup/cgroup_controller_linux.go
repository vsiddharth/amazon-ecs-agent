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
	"path/filepath"

	"github.com/aws/amazon-ecs-agent/agent/config"

	"github.com/cihub/seelog"
	"github.com/containerd/cgroups"
	"github.com/pkg/errors"
)

// validateCgroupSpec checks the cgroup spec for valid root and specifications
func validateCgroupSpec(cgroupSpec *Spec) error {
	if cgroupSpec == nil {
		return errors.New("cgroup spec validator: empty cgroup spec")
	}

	if cgroupSpec.Root == "" {
		return errors.New("cgroup spec validator: invalid cgroup root")
	}

	if !filepath.HasPrefix(cgroupSpec.Root, config.DefaultTaskCgroupPrefix) {
		return errors.New("cgroup spec validator: missing root ECS cgroup prefix")
	}

	// Validate the linux resource specs
	if cgroupSpec.Specs == nil {
		return errors.New("cgroup spec validator: empty linux resource spec")
	}
	return nil
}

// Create creates a new cgroup based off the spec post validation
func Create(cgroupSpec *Spec) error {
	seelog.Debugf("Creating cgroup %s", cgroupSpec.Root)

	// Validate incoming spec
	err := validateCgroupSpec(cgroupSpec)
	if err != nil {
		return errors.Wrapf(err, "cgroup create: failed to validate spec")
	}

	// Create cgroup
	_, err = cgroups.New(cgroups.V1, cgroups.StaticPath(cgroupSpec.Root), cgroupSpec.Specs)

	if err != nil {
		return errors.Wrapf(err, "cgroup create: unable to create controller")
	}

	return nil
}

// load is used to load the cgroup based off the spec post validation
func load(cgroupSpec *Spec) (cgroups.Cgroup, error) {
	seelog.Debugf("Loading cgroup %s", cgroupSpec.Root)

	// Validate incoming spec
	err := validateCgroupSpec(cgroupSpec)
	if err != nil {
		return nil,
			errors.Wrapf(err, "cgroup load: failed to validate spec")
	}

	// Load cgroup
	return cgroups.Load(cgroups.V1, cgroups.StaticPath(cgroupSpec.Root))
}

// Remove is used to delete the cgroup
func Remove(cgroupSpec *Spec) error {
	seelog.Debugf("Removing cgroup %s", cgroupSpec.Root)

	control, err := load(cgroupSpec)
	if err != nil {
		return errors.Wrapf(err, "cgroup remove: unable to obtain controller")
	}

	// Delete cgroup
	return control.Delete()
}
