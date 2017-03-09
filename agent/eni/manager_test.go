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

package eni

import (
	"net"
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/eni/netlinkWrapper/mocks"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

const (
	randomDevice  = "eth1"
	randomMAC     = "00:0a:95:9d:68:16"
	invalidMAC    = "0a:1b:3c:4d:5e:6ff"
	invalidDevice = "veth1"
)

func TestEmptyStateManager(t *testing.T) {
	stateManager := newStateManager()
	assert.Empty(t, stateManager.enis)
}

func TestEmptyENIManager(t *testing.T) {
	eniManager := NewENIManager()
	enis := eniManager.GetAllENIs()
	assert.Empty(t, enis)
}

func TestAddDeviceWithMACAddress(t *testing.T) {
	stateManager := newStateManager()
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	enis := stateManager.GetAllENIs()
	assert.NotEmpty(t, enis)

	err = stateManager.addDeviceWithMACAddress(randomDevice, invalidMAC)
	assert.EqualError(t, err, invalidMACMsg)

	err = stateManager.addDeviceWithMACAddress(invalidDevice, randomMAC)
	assert.EqualError(t, err, invalidDeviceMsg)
}

func TestRemoveDeviceWithMACAddress(t *testing.T) {
	stateManager := newStateManager()
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	enis := stateManager.GetAllENIs()
	assert.NotEmpty(t, enis)
	err = stateManager.removeDeviceWithMACAddress(randomMAC)
	assert.Nil(t, err)
	enis = stateManager.GetAllENIs()
	assert.Empty(t, enis)
}

func TestRemoveDevice(t *testing.T) {
	stateManager := newStateManager()
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	enis := stateManager.GetAllENIs()
	assert.NotEmpty(t, enis)
	err = stateManager.removeDevice(randomDevice)
	assert.Nil(t, err)
	enis = stateManager.GetAllENIs()
	assert.Empty(t, enis)
}

func TestDeviceExists(t *testing.T) {
	stateManager := newStateManager()
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	exists := stateManager.deviceExists(randomMAC)
	assert.True(t, exists)

	exists = stateManager.deviceExists(invalidMAC)
	assert.False(t, exists)
}

func TestENIInitStateManager(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNetlink := mock_netlinkWrapper.NewMockNetLink(mockCtrl)
	pm, _ := net.ParseMAC(randomMAC)
	mockNetlink.EXPECT().LinkList().Return([]netlink.Link{
		&netlink.Device{
			LinkAttrs: netlink.LinkAttrs{
				HardwareAddr: pm,
				Name:         randomDevice,
			},
		},
	}, nil)
	// NOTE: Set sysfsNetDir for testing purpose
	sysfsNetDir = "."
	eniManager := newStateManager()
	eniManager.netlinkClient = mockNetlink
	eniManager.InitStateManager()
	watcherChan := make(chan fsnotify.Event, 1)
	eniManager.watcher.Events = watcherChan

	enis := eniManager.GetAllENIs()
	assert.NotEmpty(t, enis)
}

func TestENIGetMACAddress(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNetlink := mock_netlinkWrapper.NewMockNetLink(mockCtrl)
	pm, _ := net.ParseMAC(randomMAC)
	mockNetlink.EXPECT().LinkByName(randomDevice).Return(
		&netlink.Device{
			LinkAttrs: netlink.LinkAttrs{
				HardwareAddr: pm,
				Name:         randomDevice,
			},
		}, nil)
	eniManager := newStateManager()
	eniManager.netlinkClient = mockNetlink
	MACAddress, err := eniManager.getMACAddress(randomDevice)
	assert.Nil(t, err)
	assert.Equal(t, randomMAC, MACAddress)
}

func TestAddDevice(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNetlink := mock_netlinkWrapper.NewMockNetLink(mockCtrl)
	pm, _ := net.ParseMAC(randomMAC)
	mockNetlink.EXPECT().LinkByName(randomDevice).Return(
		&netlink.Device{
			LinkAttrs: netlink.LinkAttrs{
				HardwareAddr: pm,
				Name:         randomDevice,
			},
		}, nil)

	eniManager := newStateManager()
	eniManager.netlinkClient = mockNetlink

	err := eniManager.addDevice(randomDevice)
	assert.Nil(t, err)
	enis := eniManager.GetAllENIs()
	assert.NotEmpty(t, enis)

	err = eniManager.addDevice(invalidDevice)
	assert.EqualError(t, err, invalidDeviceMsg)

}

func TestMACAddressValidator(t *testing.T) {
	eniManager := newStateManager()

	macStatus := eniManager.isValidMACAddress(invalidMAC)
	assert.False(t, macStatus)

	macStatus = eniManager.isValidMACAddress(randomMAC)
	assert.True(t, macStatus)
}

func TestDeviceValidator(t *testing.T) {
	eniManager := newStateManager()

	devStatus := eniManager.isValidDevice(randomDevice, ethPrefix)
	assert.True(t, devStatus)

	devStatus = eniManager.isValidDevice(invalidDevice, ethPrefix)
	assert.False(t, devStatus)
}
