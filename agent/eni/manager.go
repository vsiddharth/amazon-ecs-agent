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
	"fmt"
	"sync"
	"time"
	"os"
	"strings"
	"context"

	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/vishvananda/netlink"

	log "github.com/cihub/seelog"
)

const (
	sysfsNetDir			= "/sys/class/net"
	eth				= "eth"
	defaultReconciliationInterval	= time.Second * 10
)

type ENIManager interface {
	//TODO: Add Helper for payload handler
	InitENIStateManager() error
	BeginENIUpdate (ctx context.Context)
}


type ENIStateManager struct {
	updateLock			sync.RWMutex
	updateIntervalTicker		*time.Ticker
	enis				map[string]string // MAC => Device-Name
	watcher				*fsnotify.Watcher
}


func NewENIManager() ENIManager {
	return &ENIStateManager{
		enis: make(map[string]string, 10),
	}
}


func (eniStateManager *ENIStateManager) InitENIStateManager() error {
	links, err := netlink.LinkList()
	if err != nil {
		log.Errorf("Error retrieving network interfaces: %s", err.Error())
		return err
	}

	for _, link := range links {
		deviceName, MACAddress := link.Attrs().Name, link.Attrs().HardwareAddr.String()
		if strings.HasPrefix(deviceName, eth) {
			eniStateManager.addDeviceWithMACAddress(deviceName, MACAddress)
		}
	}

	// Setup FSNotify Watcher
	eniStateManager.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("Error creating watcher: %s", err.Error())
		return err
	}
	// Add Watch Directory
	err = eniStateManager.watcher.Add(sysfsNetDir)
	if err != nil {
		log.Errorf("Error adding watcher: %s", err.Error())
		return err
	}

	// FSNotify Update Handler
	go func () {
		for {
			select {
			case evt := <-eniStateManager.watcher.Events:
				if evt.Op&fsnotify.Create == fsnotify.Create {
					eniStateManager.addDevice(evt.Name)
				}
				if evt.Op&fsnotify.Remove == fsnotify.Remove {
					eniStateManager.removeDevice(evt.Name)
				}
			case erx := <-eniStateManager.watcher.Errors:
				log.Debugf("FSNotify Error: %s", erx.Error())
			}
		}
	}()

	return nil
}


func (eniStateManager *ENIStateManager) BeginENIUpdate (ctx context.Context) {
	eniStateManager.performPeriodicReconciliation(ctx, defaultReconciliationInterval)
}


func (eniStateManager *ENIStateManager) performPeriodicReconciliation (ctx context.Context, updateInterval time.Duration) {
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


func (eniStateManager *ENIStateManager) reconcileENIs () {
	links, err := netlink.LinkList()
	if err != nil {
		log.Error("Error obtaining netlink linklist: %s", err.Error())
	}

	currentState := eniStateManager.buildState(links)

	// Remove non-existent interfaces first
	for mac, _ := range eniStateManager.enis {
		if _, ok := currentState[mac]; !ok {
			eniStateManager.removeDeviceWithMACAddress(mac)
		}
	}

	// Add new interfaces next
	for mac, dev := range currentState {
		if !eniStateManager.deviceExists(mac) {
			eniStateManager.addDeviceWithMACAddress(mac, dev)
		}
	}
}


// Helper Methods


func (eniStateManager *ENIStateManager) addDeviceWithMACAddress (deviceName, MACAddress string) {
	eniStateManager.updateLock.Lock()
	defer eniStateManager.updateLock.Unlock()

	eniStateManager.enis[MACAddress] = deviceName
}


func (eniStateManager *ENIStateManager) addDevice (deviceName string) {
	device := filepath.Base(deviceName)
	MACAddress, err := eniStateManager.getMACAddress(device)
	if err != nil {
		log.Errorf("Error obtaining MAC Address: %s", err.Error())
		return
	}

	eniStateManager.addDeviceWithMACAddress(device, MACAddress)
}


func (eniStateManager *ENIStateManager) removeDeviceWithMACAddress (mac string) {
	eniStateManager.updateLock.Lock()
	defer eniStateManager.updateLock.Unlock()

	log.Debugf("Removing device with MACAddress: %s", mac)
	delete(eniStateManager.enis, mac)
}


func (eniStateManager *ENIStateManager) removeDevice (deviceName string) {
	for mac, dev := range eniStateManager.enis {
		if dev == deviceName {
			eniStateManager.removeDeviceWithMACAddress(mac)
		}
	}
}


func (eniStateManager *ENIStateManager) deviceExists(device string) bool {
	eniStateManager.updateLock.RLock()
	defer eniStateManager.updateLock.RUnlock()

	if _, ok := eniStateManager.enis[device]; ok {
		return true
	}
	return false
}


func (eniStateManager *ENIStateManager) getMACAddress (dev string) (string, error) {
	var mac string

	dev = filepath.Base(dev)
	link, err := netlink.LinkByName(dev)

	if err == nil {
		mac = link.Attrs().HardwareAddr.String()
	} else {
		log.Errorf("Error fetching MAC Address: %s", err.Error())
	}
	return mac, err
}

// Helper to build state for Reconciliation
func (eniStateManager *ENIStateManager) buildState(links []netlink.Link) map[string]string {
	state := make(map[string]string, 10)

	for _, link := range links {
		deviceName, MACAddress := link.Attrs().Name, link.Attrs().HardwareAddr.String()
		if strings.HasPrefix(deviceName, eth) {
			state[MACAddress] = deviceName
		}
	}
	return state
}

// Debug Purposes: Will be culled
func (eniStateManager *ENIStateManager) dumpToFile(fileName string) {
	log.Info("Dump State to FILE")
	f, err := os.Create(fileName)
	log.Errorf("Error: %s", err.Error())
	defer f.Close()

	for mac, dev := range eniStateManager.enis {
		s := fmt.Sprintf("%s - %s\n", mac, dev)
		f.WriteString(s)
		f.Sync()
	}
}
