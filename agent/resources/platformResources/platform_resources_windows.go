// +build !linux

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

import "github.com/aws/amazon-ecs-agent/agent/api"

// platformResources to abstract task platform resources
type platformResources struct{}

// New returns a new platformResources object
func New() PlatformResources {
	return &platformResources{}
}

// Setup helps setup the platform resources
func (p *platformResources) Setup(task *api.Task) error {
	return nil
}

func (p *platformResources) Cleanup(task *api.Task) error {
	return nil
}
