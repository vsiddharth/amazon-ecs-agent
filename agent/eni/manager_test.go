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
	"math/rand"
	"net"
	"strconv"
	"sync"
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

// TestEmptyStateManager checks initialization of a new State Manager
func TestEmptyStateManager(t *testing.T) {
	stateManager := newStateManager()
	assert.Empty(t, stateManager.enis)
}

// TestEmptyENIManager checks instantiation of empty ENIManager
func TestEmptyENIManager(t *testing.T) {
	eniManager := NewENIManager()
	enis := eniManager.GetAllENIs()
	assert.Empty(t, enis)
}

// TestAddDeviceWithMACAddress checks adding devices to the ENI State Manager
func TestAddDeviceWithMACAddress(t *testing.T) {
	stateManager := newStateManager()

	// Add valid (device, MAC)
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	enis := stateManager.GetAllENIs()
	assert.NotEmpty(t, enis)

	// Add device with invalid MAC
	err = stateManager.addDeviceWithMACAddress(randomDevice, invalidMAC)
	assert.EqualError(t, err, invalidMACMsg)

	// Add invalid device with valid MAC
	err = stateManager.addDeviceWithMACAddress(invalidDevice, randomMAC)
	assert.EqualError(t, err, invalidDeviceMsg)
}

// TestRemoveDeviceWithMACAddress checks removing devices from the ENI State Manager
func TestRemoveDeviceWithMACAddress(t *testing.T) {
	stateManager := newStateManager()

	// Add valid (device, MAC)
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	enis := stateManager.GetAllENIs()
	assert.NotEmpty(t, enis)

	// Remove device from State Manager
	err = stateManager.removeDeviceWithMACAddress(randomMAC)
	assert.Nil(t, err)
	enis = stateManager.GetAllENIs()
	assert.Empty(t, enis)
}

// TestRemoveDevice checks removing devices from ENI State Manager
func TestRemoveDevice(t *testing.T) {
	stateManager := newStateManager()

	// Add valid (device, MAC)
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	enis := stateManager.GetAllENIs()
	assert.NotEmpty(t, enis)

	// Remove device from State Manager
	err = stateManager.removeDevice(randomDevice)
	assert.Nil(t, err)
	enis = stateManager.GetAllENIs()
	assert.Empty(t, enis)
}

// TestDeviceExists checks the existence of devices in State Manager
func TestDeviceExists(t *testing.T) {
	stateManager := newStateManager()

	// Add valid (device, MAC)
	err := stateManager.addDeviceWithMACAddress(randomDevice, randomMAC)
	assert.Nil(t, err)
	exists := stateManager.deviceExists(randomMAC)
	assert.True(t, exists)

	exists = stateManager.deviceExists(invalidMAC)
	assert.False(t, exists)
}

// TestENIInitStateManager checks the sanity of InitStateManager
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

	// NOTE: Set sysfsNetDir for testing purposes only
	sysfsNetDir = "."
	eniManager := newStateManager()
	eniManager.netlinkClient = mockNetlink
	eniManager.InitStateManager()
	watcherChan := make(chan fsnotify.Event, 1)
	eniManager.watcher.Events = watcherChan

	enis := eniManager.GetAllENIs()
	assert.NotEmpty(t, enis)
}

// TestENIGetMACAddress checks getMACAddress
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

// TestAddDevice checks adding devices to the ENI State Manager
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

	// Add valid device to State Manager
	err := eniManager.addDevice(randomDevice)
	assert.Nil(t, err)
	enis := eniManager.GetAllENIs()
	assert.NotEmpty(t, enis)

	// Attempt to add an invalid device
	err = eniManager.addDevice(invalidDevice)
	assert.EqualError(t, err, invalidDeviceMsg)
}

// TestMACAddressValidator verifies MAC address added to State Manager
func TestMACAddressValidator(t *testing.T) {
	eniManager := newStateManager()

	macStatus := eniManager.isValidMACAddress(invalidMAC)
	assert.False(t, macStatus)

	macStatus = eniManager.isValidMACAddress(randomMAC)
	assert.True(t, macStatus)
}

// TestDeviceValidator verifies valid device names
func TestDeviceValidator(t *testing.T) {
	eniManager := newStateManager()

	devStatus := eniManager.isValidDevice(randomDevice, ethPrefix)
	assert.True(t, devStatus)

	devStatus = eniManager.isValidDevice(invalidDevice, ethPrefix)
	assert.False(t, devStatus)
}

// Generate Random MAC Address
func genRandomMACAddress() string {
	validAlphabet := "0123456789ABCDEF"
	lmac := 12
	b := make([]byte, lmac)

	for i := range b {
		b[i] = validAlphabet[rand.Intn(len(validAlphabet))]
	}

	mac := string(b)
	for i := 2; i < len(mac); i += 3 {
		mac = mac[:i] + ":" + mac[i:]
	}
	return mac

}

// TestConcurrentAddDevice checks concurrent state updates
func TestConcurrentAddDevice(t *testing.T) {
	var waitGroup sync.WaitGroup
	numRountines := 8000

	eniManager := newStateManager()

	waitGroup.Add(numRountines)

	for i := 0; i < numRountines; i++ {
		dev := ethPrefix + strconv.Itoa(i)
		mac := genRandomMACAddress()
		go func() {
			eniManager.updateLock.Lock()
			eniManager.addDeviceWithMACAddress(dev, mac)
			eniManager.updateLock.Unlock()
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()

	enis := eniManager.GetAllENIs()
	assert.Equal(t, len(enis), numRountines)
}

// TestConcurrentRemoveDevice checks concurrent state updates
func TestConcurrentRemoveDevice(t *testing.T) {
	var waitGroup sync.WaitGroup
	numRountines := 80

	eniManager := newStateManager()

	for i := 0; i < numRountines; i++ {
		dev := ethPrefix + strconv.Itoa(i)
		mac := genRandomMACAddress()
		eniManager.updateLock.Lock()
		eniManager.addDeviceWithMACAddress(dev, mac)
		eniManager.updateLock.Unlock()
	}

	waitGroup.Add(numRountines)

	for i := 0; i < numRountines; i++ {
		dev := ethPrefix + strconv.Itoa(i)
		go func() {
			eniManager.updateLock.Lock()
			eniManager.removeDevice(dev)
			eniManager.updateLock.Unlock()
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()

	enis := eniManager.GetAllENIs()
	assert.Equal(t, len(enis), 0)
}
