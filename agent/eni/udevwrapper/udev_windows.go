// +build windows

// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package udevwrapper

import (
	"github.com/aws/amazon-ecs-agent/agent/eni/enimonitor"
)

// Udev Wrapper methods used from the  package
type Udev interface {
	Monitor(notify chan *enimonitor.ENIEvent) (shutdown chan bool)
	Close() error
}

// New returns an UDev Monitor
func New() (*enimonitor.ENIMonitor, error) {
	return enimonitor.NewMonitor()
}
