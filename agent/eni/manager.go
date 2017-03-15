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
	"context"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/aws/amazon-ecs-agent/agent/eni/netlinkWrapper"
	log "github.com/cihub/seelog"
)

const (
	ethPrefix                     = "eth"
	defaultReconciliationInterval = time.Second * 30
	invalidDeviceMsg              = "Invalid Device Name"
	invalidMACMsg                 = "Invalid MAC Address"
)

var sysfsNetDir = "/sys/class/net"

// Manager exposes the methods to initialize and update ENI's
// attached to the instance.
type Manager interface {
	InitStateManager() error
	BeginENIUpdate(ctx context.Context)
	GetAllENIs() map[string]string
}

// StateManager maintains the state of ENI's connected
// to the instance. It also has supporting elements to
// maintain consistency and update intervals
type StateManager struct {
	updateLock           sync.RWMutex
	updateIntervalTicker *time.Ticker
	enis                 map[string]string // MAC => Device-Name
	watcher              *fsnotify.Watcher
	netlinkClient        netlinkWrapper.NetLink
}

// NewENIManager instanciates a new ENIStateManager
func NewENIManager() Manager {
	return newStateManager()
}

func newStateManager() *StateManager {
	return &StateManager{
		enis:          make(map[string]string, 10),
		netlinkClient: netlinkWrapper.NetLinkClient{},
	}
}

// InitStateManager initializes a new ENI State Manager
func (eniStateManager *StateManager) InitStateManager() error {
	links, err := eniStateManager.netlinkClient.LinkList()
	if err != nil {
		log.Errorf("Error retrieving network interfaces: %v", err)
		return err
	}

	eniStateManager.updateLock.Lock()
	for _, link := range links {
		deviceName, MACAddress := link.Attrs().Name, link.Attrs().HardwareAddr.String()
		if strings.HasPrefix(deviceName, ethPrefix) {
			err = eniStateManager.addDeviceWithMACAddress(deviceName, MACAddress)
			if err != nil {
				log.Errorf(err.Error())
			}
		}
	}
	eniStateManager.updateLock.Unlock()

	// Setup FSNotify Watcher
	eniStateManager.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("Error creating watcher: %v", err)
		return err
	}
	// Add Watch Directory
	err = eniStateManager.watcher.Add(sysfsNetDir)
	if err != nil {
		log.Errorf("Error adding watcher: %v", err)
		return err
	}

	// FSNotify Update Handler
	go eniStateManager.fsnotifyHandler()

	return nil
}

// BeginENIUpdate periodically updates the state of ENI's connected to the system
func (eniStateManager *StateManager) BeginENIUpdate(ctx context.Context) {
	eniStateManager.performPeriodicReconciliation(ctx, defaultReconciliationInterval)
}

func (eniStateManager *StateManager) performPeriodicReconciliation(ctx context.Context, updateInterval time.Duration) {
	eniStateManager.updateIntervalTicker = time.NewTicker(updateInterval)
	for {
		select {
		case <-eniStateManager.updateIntervalTicker.C:
			go eniStateManager.reconcileENIs()
		case <-ctx.Done():
			eniStateManager.updateIntervalTicker.Stop()
			return
		}
	}
}

func (eniStateManager *StateManager) reconcileENIs() {
	links, err := eniStateManager.netlinkClient.LinkList()
	if err != nil {
		log.Errorf("Error obtaining netlink linklist: %v", err)
	}

	currentState := eniStateManager.buildState(links)

	// Remove non-existent interfaces first
	eniStateManager.updateLock.Lock()
	for mac := range eniStateManager.enis {
		if _, ok := currentState[mac]; !ok {
			err = eniStateManager.removeDeviceWithMACAddress(mac)
			if err != nil {
				log.Errorf(err.Error())
			}
		}
	}
	eniStateManager.updateLock.Unlock()

	// Add new interfaces next
	for mac, dev := range currentState {
		if !eniStateManager.deviceExists(mac) {
			eniStateManager.updateLock.Lock()
			err = eniStateManager.addDeviceWithMACAddress(dev, mac)
			if err != nil {
				log.Errorf(err.Error())
			}
			eniStateManager.updateLock.Unlock()
		}
	}
}

func (eniStateManager *StateManager) GetAllENIs() map[string]string {
	return eniStateManager.enis
}

// Helper Methods

//NOTE: addDeviceWithMACAddress expects lock to be held prior to update
func (eniStateManager *StateManager) addDeviceWithMACAddress(deviceName, MACAddress string) error {
	log.Debugf("Adding device %s with MAC %s", deviceName, MACAddress)

	// Validate parameters for correctness
	if !eniStateManager.isValidDevice(deviceName, ethPrefix) {
		return errors.New(invalidDeviceMsg)
	}

	if !eniStateManager.isValidMACAddress(MACAddress) {
		return errors.New(invalidMACMsg)
	}

	eniStateManager.enis[MACAddress] = deviceName
	return nil
}

func (eniStateManager *StateManager) addDevice(deviceName string) error {
	device := filepath.Base(deviceName)

	if !eniStateManager.isValidDevice(deviceName, ethPrefix) {
		return errors.New(invalidDeviceMsg)
	}

	MACAddress, err := eniStateManager.getMACAddress(device)

	if err != nil {
		log.Errorf("Error obtaining MAC Address: %v", err)
		return err
	}

	return eniStateManager.addDeviceWithMACAddress(device, MACAddress)
}

//NOTE: removeDeviceWithMACAddress expects lock to be held prior to update
func (eniStateManager *StateManager) removeDeviceWithMACAddress(mac string) error {
	log.Debugf("Removing device with MACAddress: %s", mac)

	if !eniStateManager.isValidMACAddress(mac) {
		return errors.New(invalidMACMsg)
	}

	delete(eniStateManager.enis, mac)
	return nil
}

func (eniStateManager *StateManager) removeDevice(deviceName string) error {
	log.Debugf("Removing device: %s", deviceName)

	if !eniStateManager.isValidDevice(deviceName, ethPrefix) {
		return errors.New(invalidDeviceMsg)
	}

	for mac, dev := range eniStateManager.enis {
		if dev == deviceName {
			eniStateManager.removeDeviceWithMACAddress(mac)
		}
	}
	return nil
}

func (eniStateManager *StateManager) deviceExists(mac string) bool {
	eniStateManager.updateLock.RLock()
	defer eniStateManager.updateLock.RUnlock()

	if _, ok := eniStateManager.enis[mac]; ok {
		return true
	}
	return false
}

func (eniStateManager *StateManager) getMACAddress(dev string) (string, error) {
	var mac string

	dev = filepath.Base(dev)
	link, err := eniStateManager.netlinkClient.LinkByName(dev)

	if err == nil {
		mac = link.Attrs().HardwareAddr.String()
	}
	return mac, err
}

func (eniStateManager *StateManager) isValidDevice(deviceName, prefix string) bool {
	if strings.HasPrefix(deviceName, prefix) {
		return true
	}
	return false
}

func (eniStateManager *StateManager) isValidMACAddress(mac string) bool {
	_, err := net.ParseMAC(mac)
	if err != nil {
		return false
	}
	return true
}

// Helper to build state for Reconciliation
func (eniStateManager *StateManager) buildState(links []netlink.Link) map[string]string {
	state := make(map[string]string, 10)

	for _, link := range links {
		deviceName, MACAddress := link.Attrs().Name, link.Attrs().HardwareAddr.String()
		if strings.HasPrefix(deviceName, ethPrefix) {
			state[MACAddress] = deviceName
		}
	}
	return state
}

func (eniStateManager *StateManager) fsnotifyHandler() {
	for {
		select {
		case evt := <-eniStateManager.watcher.Events:
			if evt.Op&fsnotify.Create == fsnotify.Create {
				eniStateManager.updateLock.Lock()
				eniStateManager.addDevice(evt.Name)
				eniStateManager.updateLock.Unlock()
			}
			if evt.Op&fsnotify.Remove == fsnotify.Remove {
				eniStateManager.updateLock.Lock()
				eniStateManager.removeDevice(evt.Name)
				eniStateManager.updateLock.Unlock()
			}
		case erx := <-eniStateManager.watcher.Errors:
			log.Debugf("FSNotify Error: %s", erx.Error())
		}
	}
}
