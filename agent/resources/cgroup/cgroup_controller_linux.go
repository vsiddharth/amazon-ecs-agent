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
	"github.com/cihub/seelog"
	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

var factory CgroupFactory = &GlobalCgroupFactory{}

// Create creates a new cgroup based off the spec post validation
func Create(cgroupSpec *Spec) (cgroups.Cgroup, error) {
	// Validate incoming spec
	err := validateCgroupSpec(cgroupSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "cgroup create: failed to validate spec")
	}

	// Create cgroup
	seelog.Infof("Creating cgroup %s", cgroupSpec.Root)
	control, err := factory.New(cgroups.V1, cgroups.StaticPath(cgroupSpec.Root), cgroupSpec.Specs)

	if err != nil {
		return nil, errors.Wrapf(err, "cgroup create: unable to create controller")
	}

	return control, nil
}

// Remove is used to delete the cgroup
func Remove(cgroupPath string) error {
	seelog.Debugf("Removing cgroup %s", cgroupPath)

	control, err := factory.Load(cgroups.V1, cgroups.StaticPath(cgroupPath))
	if err != nil {
		return errors.Wrapf(err, "cgroup remove: unable to obtain controller")
	}

	// Delete cgroup
	return control.Delete()
}

// validateCgroupSpec checks the cgroup spec for valid path and specifications
func validateCgroupSpec(cgroupSpec *Spec) error {
	if cgroupSpec == nil {
		return errors.New("cgroup spec validator: empty cgroup spec")
	}

	if cgroupSpec.Root == "" {
		return errors.New("cgroup spec validator: invalid cgroup root")
	}

	// Validate the linux resource specs
	if cgroupSpec.Specs == nil {
		return errors.New("cgroup spec validator: empty linux resource spec")
	}
	return nil
}

//go:generate go run ../../../scripts/generate/mockgen.go github.com/containerd/cgroups Cgroup mock/cgroups.go
type CgroupFactory interface {
	New(hierarchy cgroups.Hierarchy, path cgroups.Path, specs *specs.LinuxResources) (cgroups.Cgroup, error)
	Load(hierarchy cgroups.Hierarchy, path cgroups.Path) (cgroups.Cgroup, error)
}

// GlobalCgroupFactory calls the cgroups library global functions
type GlobalCgroupFactory struct{}

func (c *GlobalCgroupFactory) Load(hierarchy cgroups.Hierarchy, path cgroups.Path) (cgroups.Cgroup, error) {
	return cgroups.Load(hierarchy, path)
}

func (c *GlobalCgroupFactory) New(hierarchy cgroups.Hierarchy, path cgroups.Path, specs *specs.LinuxResources) (cgroups.Cgroup, error) {
	return cgroups.New(hierarchy, path, specs)
}
