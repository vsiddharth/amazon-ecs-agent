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
	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Spec captures the abstraction for a creating a new
// cgroup based on root and the runtime specifications
type Spec struct {
	// Root is the cgroup path
	Root string
	// Specs are for all the linux resources including cpu, memory, etc...
	Specs *specs.LinuxResources
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
