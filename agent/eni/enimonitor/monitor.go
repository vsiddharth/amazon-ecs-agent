package enimonitor

import (
	"net"
	"syscall"
	"unsafe"
	log "github.com/cihub/seelog"
	"golang.org/x/sys/windows"
)

// UDevMonitor is a UDev netlink socket
type ENIMonitor struct {
	procWSACreateEvent *windows.LazyProc
	procNotifyAddrChange *windows.LazyProc
	overlap *windows.Overlapped
}

type ENIEvent struct {
	Index        int          // positive integer that starts at one, zero is never used
	MTU          int          // maximum transmission unit
	Name         string       // e.g., "en0", "lo0", "eth0.100"
	HardwareAddr string // IEEE MAC-48, EUI-48 and EUI-64 form
	Flags        string        // e.g., FlagUp, FlagLoopback, FlagMulticast
}

// NewMonitor creates and connects a new monitor
func NewMonitor() (mon *ENIMonitor, err error) {
	mon = new(ENIMonitor)

	var modws232 = windows.NewLazySystemDLL("ws2_32.dll")
	var modiphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")

	mon.procWSACreateEvent  = modws232.NewProc("WSACreateEvent")
	mon.procNotifyAddrChange = modiphlpapi.NewProc("NotifyAddrChange")

	mon.overlap = &windows.Overlapped{}
	mon.overlap.HEvent, err = mon.WSACreateEvent()
	return
}


// Close closes the monitor socket
func (mon *ENIMonitor) Close() error {
	return windows.Close(mon.overlap.HEvent)
}

// Process processes one packet from the socket, and sends the event on the notify channel
func (mon *ENIMonitor) Process(notify chan *ENIEvent) {

	for {
		log.Debugf("Invoking NotifyAddrChange()")
		notifyHandle := windows.Handle(0)
		syscall.Syscall(uintptr(mon.procNotifyAddrChange.Addr()), 2, uintptr(notifyHandle), uintptr(unsafe.Pointer(mon.overlap)), 0)

		log.Debugf("Waiting for network changes")
		event, err := windows.WaitForSingleObject(mon.overlap.HEvent, windows.INFINITE)

		if err != nil {
			log.Errorf("Error occurred while waiting for windows network address change event")
		}

		switch event {
		case windows.WAIT_OBJECT_0:
			log.Debugf("Windows kernel notified of a network address change")
			l, err := net.Interfaces()
			if err != nil {
				panic(err)
			}

			for _, f := range l {
				//Take only up and ignore loopback interfaces
				if (f.Flags & net.FlagUp != 0) && (f.Flags & net.FlagLoopback == 0) {
					event := &ENIEvent{
						Index:        f.Index,
						MTU:          f.MTU,
						Name:         f.Name,
						HardwareAddr: f.HardwareAddr.String(),
						Flags:        f.Flags.String(),
					}

					notify <- event
				}
			}

		default:
			break
		}
	}
}

// Monitor starts udev event monitoring. Events are sent on the notify channel, and the watch can be
// terminated by sending true on the returned shutdown channel
func (mon *ENIMonitor) Monitor(notify chan *ENIEvent) (shutdown chan bool) {
	shutdown = make(chan bool)

	go func() {
	done:
		for {
			select {
			case <-shutdown:
				break done
			default:
				mon.Process(notify)
			}
		}
	}()
	return shutdown
}

func (mon *ENIMonitor) WSACreateEvent() (windows.Handle, error) {
	handlePtr, _, errNum := syscall.Syscall(mon.procWSACreateEvent.Addr(), 0, 0, 0, 0)
	if handlePtr == 0 {
		return 0, errNum
	} else {
		return windows.Handle(handlePtr), nil
	}
}